/* Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"fmt"

	log "github.com/golang/glog"
)

// configEthIntf configures the network interfaces on this device based on the given configuration.
func configEthIntf(ethIntfName string, vlanIDs []int) error {
	log.Infof("Configuring interface %v. VLAN: %v.", ethIntfName, vlanIDs)
	vlanIntfNames := make(map[int]string) // VLAN ID -> VLAN Intf name
	for _, vlanID := range vlanIDs {
		// Add VLAN interface
		vlanIntfName, err := cmdRunner.CreateVLAN(ethIntfName, vlanID)
		if err != nil {
			return err
		}
		vlanIntfNames[vlanID] = vlanIntfName
	}

	// Restart eth interface.
	if err := cmdRunner.RestartIntf(ethIntfName); err != nil {
		return err
	}

	for vlanID, vlanIntfName := range vlanIntfNames {
		// Wipe out IP on VLAN interface.
		if err := cmdRunner.WipeOutIntfIP(vlanIntfName); err != nil {
			return err
		}

		// Add a network bridge
		bridgeName := getBridgeName(vlanID)
		if err := cmdRunner.CreateBridge(bridgeName); err != nil {
			return err
		}

		// Link VLAN intf to bridge
		if err := cmdRunner.AddBridgeIntf(bridgeName, vlanIntfName); err != nil {
			return err
		}
	}

	log.Infof("Configured interface %v.", ethIntfName)
	return nil
}

// cleanupEthIntf cleans up the network interfaces on this device based on the given configuration.
// It goes through all cleanup steps even if some failures are detected, and returns all errors.
func cleanupEthIntf(ethIntfName string, vlanIDs []int) []error {
	log.Infof("Cleaning up interface %s. VLAN: %d.", ethIntfName, vlanIDs)
	var errs []error

	for _, vlanID := range vlanIDs {
		// Delete VLAN interface.
		if err := cmdRunner.DeleteVLAN(ethIntfName, vlanID); err != nil {
			errs = append(errs, err)
		}

		// Turn down network bridge.
		bridgeName := getBridgeName(vlanID)
		if err := cmdRunner.TurnDownIntf(bridgeName); err != nil {
			errs = append(errs, err)
		}

		// Remove network bridge.
		if err := cmdRunner.DeleteBridge(bridgeName); err != nil {
			errs = append(errs, err)
		}
	}

	log.Infof("Cleaned up interface %s. Number of errors = %d.", ethIntfName, len(errs))
	return errs
}

// configWLANIntf configures the network interfaces to make it work with hostapd.
func configWLANIntf(wlanIntfName string) error {
	log.Infof("Configuring WLAN interface %v.", wlanIntfName)

	// Wipe out IP on WLAN interface.
	if err := cmdRunner.WipeOutIntfIP(wlanIntfName); err != nil {
		return err
	}

	if err := cmdRunner.TurnDownIntf(wlanIntfName); err != nil {
		return err
	}

	wlanIntfMAC, err := cmdRunner.IntfMAC(wlanIntfName)
	if err != nil {
		return err
	}

	// Change the MAC address of the wireless interface to avoid conflict with other devices.
	updatedMAC := generateWLANIntfMAC(wlanIntfMAC)
	if err = cmdRunner.UpdateIntfMAC(wlanIntfName, updatedMAC); err != nil {
		return err
	}

	if err = cmdRunner.BringUpIntf(wlanIntfName); err != nil {
		return err
	}

	log.Infof("Configured WLAN interface %v.", wlanIntfName)
	return nil
}

func getBridgeName(vlanID int) string {
	return fmt.Sprintf("br_%d", vlanID)
}

func generateWLANIntfMAC(originalMAC string) string {
	return fmt.Sprintf("02%s0", originalMAC[2:len(originalMAC)-1])
}
