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

package syscmd

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/golang/glog"
)

// CreateVLAN creates vlan with a specific ID on target interface.
// It returns created interface name if succeeded, or error if failed.
func (r *CommandRunner) CreateVLAN(intfName string, vlanID int) (string, error) {
	vlanINTFName := vlanINTFName(intfName, vlanID)
	log.Infof("Creating VLAN interface %v...", vlanINTFName)
	if _, err := r.ExecCommand(true, "ip", "link", "add", "link", intfName, "name", vlanINTFName, "type", "vlan", "id", strconv.Itoa(vlanID)); err != nil {
		return "", err
	}
	log.Infof("Created VLAN interface %v.", vlanINTFName)
	return vlanINTFName, nil
}

// DeleteVLAN deletes the vlan with a specific ID on target interface.
func (r *CommandRunner) DeleteVLAN(intfName string, vlanID int) error {
	vlanINTFName := vlanINTFName(intfName, vlanID)
	log.Infof("Deleting VLAN interface %v...", vlanINTFName)
	if _, err := r.ExecCommand(true, "ip", "link", "delete", vlanINTFName); err != nil {
		return err
	}
	log.Infof("Deleted VLAN interface %v.", vlanINTFName)
	return nil
}

// RestartIntf restarts a specific network interface.
func (r *CommandRunner) RestartIntf(intfName string) error {
	log.Infof("Restarting interface %v...", intfName)

	if err := r.TurnDownIntf(intfName); err != nil {
		return err
	}

	if err := r.BringUpIntf(intfName); err != nil {
		return err
	}

	log.Infof("Restarted the interface %v.", intfName)
	return nil
}

// BringUpIntf brings up a certain network interface.
func (r *CommandRunner) BringUpIntf(intfName string) error {
	if _, err := r.ExecCommand(true, "ifconfig", intfName, "up"); err != nil {
		return err
	}
	log.Infof("Interface %v is UP.", intfName)
	return nil
}

// TurnDownIntf turns down a certain network interface.
func (r *CommandRunner) TurnDownIntf(intfName string) error {
	if _, err := r.ExecCommand(true, "ifconfig", intfName, "down"); err != nil {
		return err
	}
	log.Infof("Interface %v is DOWN.", intfName)
	return nil
}

// WipeOutIntfIP cleans up the IP address on a certain network interface.
func (r *CommandRunner) WipeOutIntfIP(intfName string) error {
	if _, err := r.ExecCommand(true, "ifconfig", intfName, "0.0.0.0"); err != nil {
		return err
	}
	log.Infof("Wiped out the IP on interface %v.", intfName)
	return nil
}

// IntfMac returns the MAC address of a certain interface.
func (r *CommandRunner) IntfMAC(intfName string) (string, error) {
	mac, err := r.ExecCommand(true, "cat", fmt.Sprintf("/sys/class/net/%s/address", intfName))
	if err != nil {
		return "", err
	}
	mac = mac[0 : len(mac)-1] // The last charactor is '\n'
	log.Infof("MAC address of %v is %v.", intfName, mac)
	return mac, nil
}

// VLANOnIntf returns IDs of all VLAN on the given interface.
func (r *CommandRunner) VLANOnIntf(intfName string) ([]int, error) {
	// Fetch all interface information on the device.
	linkInfo, err := r.ExecCommand(true, "ip", "-o", "-d", "link", "show")
	if err != nil {
		return nil, err
	}

	// Parse the result to find the target VLAN interface.
	vlanIDs := vlanIDsInIPLinkResult(intfName, linkInfo)
	log.Infof("Interface %s has VLAN %v.", intfName, vlanIDs)
	return vlanIDs, nil
}

func vlanIDsInIPLinkResult(intfName, linkInfo string) []int {
	var vlanIDs []int
	for _, intfInfo := range strings.Split(linkInfo, "\n") {
		if !strings.Contains(intfInfo, fmt.Sprintf("@%s", intfName)) || !strings.Contains(intfInfo, "vlan") {
			continue
		}
		for _, infoElem := range strings.Split(intfInfo, "\\") {
			if !strings.Contains(infoElem, "vlan") {
				continue
			}
			foundID := false
			for _, vlanInfoElem := range strings.Fields(infoElem) {
				if foundID {
					if vlanID, err := strconv.Atoi(vlanInfoElem); err == nil {
						vlanIDs = append(vlanIDs, vlanID)
					}
					break
				}
				if vlanInfoElem == "id" {
					foundID = true
				}
			}
		}
	}
	return vlanIDs
}

// UpdateIntfMAC changes the MAC address of a certain interface to the inputed one.
func (r *CommandRunner) UpdateIntfMAC(intfName, updatedMAC string) error {
	if _, err := r.ExecCommand(true, "ifconfig", intfName, "hw", "ether", updatedMAC); err != nil {
		return err
	}
	log.Infof("The MAC address of %v updated to %v.", intfName, updatedMAC)
	return nil
}

// SendDHCPRequest sends a DHCP request for a certain network interface with the hostname in parameter.
func (r *CommandRunner) SendDHCPRequest(intfName, hostname string) error {
	if _, err := r.ExecCommand(false, "udhcpc", "-i", intfName, "-x", fmt.Sprintf("hostname:%s", hostname)); err != nil {
		return err
	}
	log.Infof("Send DHCP request on interface %s with hostname %s.", intfName, hostname)
	return nil
}
