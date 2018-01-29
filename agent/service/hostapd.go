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
	"errors"
	"fmt"
	"path"

	log "github.com/golang/glog"
	"github.com/google/link022/agent/syscmd"
	"github.com/google/link022/agent/util/ocutil"
	"github.com/google/link022/generated/ocstruct"
)

const (
	ctrlInterfaceConfigTemplate = `ctrl_interface=%s
`
	radiusAttributeSaveConfigTemplate = `radius_auth_access_accept_attr=%s
`
	commonConfigTemplate = `
interface=%s
# Driver; nl80211 is used with all Linux mac80211 drivers.
driver=nl80211
hw_mode=%s
channel=%d

`

	bssConfigTemplate = `
# bssid for multiple wlans, the format is like "wlan0_1"
# For the first wlan, there should be no bssid field, otherwise hostapd
# will fail to start.
bss=%s_%d
`

	wlanConfigTemplate = `ssid=%s
bridge=%s
ap_isolate=%d
`

	authConfigTemplate = `ieee8021x=1
auth_algs=1
wpa=2
rsn_pairwise=CCMP
wpa_key_mgmt=WPA-EAP
macaddr_acl=0
auth_server_addr=%s
auth_server_port=%d
auth_server_shared_secret=%s
nas_identifier=%s
`
)

// configHostapd configures the hostapd program on this device based on the given AP configuration.
func configHostapd(apConfig *ocstruct.Device, wlanINTFName string) error {
	hostname := *apConfig.Hostname
	apRadios := apConfig.Radios
	ctrlInterface := ""
	if apConfig.VendorConfig["ctrl_interface"] != nil {
		ctrlInterface = *apConfig.VendorConfig["ctrl_interface"].ConfigValue
	}
	radiusAttribute := ""
	if apConfig.VendorConfig["radius_auth_access_accept_attr"] != nil {
		radiusAttribute = *apConfig.VendorConfig["radius_auth_access_accept_attr"].ConfigValue
	}
	if apRadios == nil || len(apRadios.Radio) == 0 {
		log.Error("No radio configuration found.")
		return errors.New("no radio configuration found")
	}

	if len(apRadios.Radio) > 1 {
		log.Errorf("Invalid radio number, expected: 1, actual: %d.", len(apRadios.Radio))
		return errors.New("not supporting multiple radios")
	}

	authServerConfigs := ocutil.RadiusServers(apConfig)
	for _, apRadio := range apRadios.Radio {
		radioConfig := apRadio.Config
		wlanConfigs := wlanWithOpFreq(apConfig, radioConfig.OperatingFrequency)

		// Genearte hostapd configuration.
		hostapdConfig := hostapdConfigFile(radioConfig, authServerConfigs, wlanConfigs, wlanINTFName, hostname, ctrlInterface, radiusAttribute)


		// Save the hostapd configuration file.
		configFileName := hostapdConfFileName(wlanINTFName)
		if err := syscmd.SaveToFile(runFolder, configFileName, hostapdConfig); err != nil {
			return err
		}

		// Start hostapd.
		if err := cmdRunner.StartHostapd(path.Join(runFolder, configFileName)); err != nil {
			return err
		}
	}

	return nil
}

// hostapdConfigFile generates the content of hostapd configuration file based on the given configuration.
func hostapdConfigFile(radioConfig *ocstruct.OpenconfigOfficeAp_Radios_Radio_Config,
	authServerConfigs map[string]*ocstruct.OpenconfigOfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server,
	wlanConfigs []*ocstruct.OpenconfigOfficeAp_Ssids_Ssid_Config,
	wlanINTFName string, hostname string, ctrlInterface string, radiusAttribute string) string {
	log.Infof("Generating hostapd configuration for radio %v...", *radioConfig.Id)
	hostapdConfig := ""

	// Generate common configuration.
	radioHWMode := hostapdHardwareMode(radioConfig.OperatingFrequency)
	commonConfig := fmt.Sprintf(commonConfigTemplate, wlanINTFName, radioHWMode, *radioConfig.Channel)

	if len(ctrlInterface) != 0 {
		commonConfig += fmt.Sprintf(ctrlInterfaceConfigTemplate, ctrlInterface)
	}
	if len(radiusAttribute) != 0 {
		commonConfig += fmt.Sprintf(radiusAttributeSaveConfigTemplate, radiusAttribute)
	}

	hostapdConfig += commonConfig

	// Generate wlan configuration.
	for i, wlanConfig := range wlanConfigs {
		wlanName := *wlanConfig.Name
		log.Infof("Adding hostapd configuration for WLAN %v...", wlanName)

		if i > 0 {
			// Add BSS configuration.
			bssConfig := fmt.Sprintf(bssConfigTemplate, wlanINTFName, i)
			hostapdConfig += bssConfig
		}

		// Add WLAN configuration.
		wlanBridgeName := getBridgeName(int(*wlanConfig.VlanId))

		wlanStationIsolation := 0
		if wlanConfig.StationIsolation != nil && *wlanConfig.StationIsolation{
				wlanStationIsolation = 1
		}

		hostapdWLANConfig := fmt.Sprintf(wlanConfigTemplate, wlanName, wlanBridgeName, wlanStationIsolation)
		hostapdConfig += hostapdWLANConfig

		// Add AUTH configuration.
		if wlanConfig.Opmode == ocstruct.OpenconfigOfficeAp_Ssids_Ssid_Config_Opmode_WPA2_ENTERPRISE {
			// Add radius configuration.
			authServerConfig := authServerConfigs[wlanName]
			// TODO: Add validation to ensure authServerConfig exists.
			radiusServerAddr := *authServerConfig.Address
			radiusServerPort := *authServerConfig.Radius.Config.AuthPort
			radiusSecret := *authServerConfig.Radius.Config.SecretKey
			authConfig := fmt.Sprintf(authConfigTemplate, radiusServerAddr, radiusServerPort, radiusSecret, hostname)
			hostapdConfig += authConfig
		}
		// TODO: Add validation to block WPA2_PERSONAL.
	}

	log.Info("Generated hostapd configuration.")
	return hostapdConfig
}

func hostapdHardwareMode(opFrequency ocstruct.E_OpenconfigWifiTypes_OPERATING_FREQUENCY) string {
	if opFrequency == ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2GHZ ||
		opFrequency == ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2_5_GHZ {
		return "g"
	}
	return "a"
}

func wlanWithOpFreq(apConfig *ocstruct.Device,
	targetFreq ocstruct.E_OpenconfigWifiTypes_OPERATING_FREQUENCY) []*ocstruct.OpenconfigOfficeAp_Ssids_Ssid_Config {
	var matchedWLANs []*ocstruct.OpenconfigOfficeAp_Ssids_Ssid_Config

	wlans := apConfig.Ssids
	if wlans == nil || len(wlans.Ssid) == 0 {
		// No WLAN on this AP.
		return matchedWLANs
	}

	for _, wlan := range wlans.Ssid {
		wlanConfig := wlan.Config
		if wlanConfig.OperatingFrequency == ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2_5_GHZ ||
			wlanConfig.OperatingFrequency == targetFreq {
			matchedWLANs = append(matchedWLANs, wlanConfig)
		}
	}
	return matchedWLANs
}
func hostapdConfFileName(wlanINTFName string) string {
	return fmt.Sprintf("hostapd_%s.conf", wlanINTFName)
}
