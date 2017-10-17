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

// The openconfig program contains an example demonstrating how to use
// the auto-generated wireless openconfig module.
package main

import (
	log "github.com/golang/glog"
	"github.com/google/link022/generated/ocstruct"
	"github.com/openconfig/ygot/ygot"
)

const (
	officeName = "emulator-office"

	ap1Name = "AP-1"
	ap2Name = "AP-2"

	vendor = "link022"

	radioID = 1

	guestWLANName = "Guest-Emu"
	authWLANName  = "Auth-Emu"

	orgIDConfigKey = "org_id"
	orgID          = "xxxxx-xxxx-xxxx-xxxx-xxxxxxxxx"

	usernameConfigKey = "user_name"
	username          = "admin"
)

func addAPs(wifiOffice *ocstruct.Office, apNum int, addAuthWLAN bool) {
	if apNum <= 0 {
		return
	}

	wifiOffice.OfficeAp = make(map[string]*ocstruct.WifiOffice_OfficeAp)

	// Add AP 1.
	ap1 := &ocstruct.WifiOffice_OfficeAp{
		Hostname: ygot.String(ap1Name),
		Vendor:   ygot.String(vendor),
	}
	wifiOffice.OfficeAp[ap1Name] = ap1
	addRadios(ap1)
	addWLANs(ap1, addAuthWLAN)

	if apNum <= 1 {
		return
	}

	// Add AP 2.
	ap2 := &ocstruct.WifiOffice_OfficeAp{
		Hostname: ygot.String(ap2Name),
		Vendor:   ygot.String(vendor),
	}
	wifiOffice.OfficeAp[ap2Name] = ap2
	addRadios(ap2)
	addWLANs(ap2, addAuthWLAN)
}

func addRadios(ap *ocstruct.WifiOffice_OfficeAp) {
	radios := &ocstruct.WifiOffice_OfficeAp_Radios{}
	radios.Radio = make(map[uint8]*ocstruct.WifiOffice_OfficeAp_Radios_Radio)
	ap.Radios = radios

	radios.Radio[radioID] = &ocstruct.WifiOffice_OfficeAp_Radios_Radio{
		Id: ygot.Uint8(radioID),
		Config: &ocstruct.WifiOffice_OfficeAp_Radios_Radio_Config{
			Id:                 ygot.Uint8(radioID),
			Enabled:            ygot.Bool(true),
			OperatingFrequency: ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2GHZ,
			TransmitPower:      ygot.Uint8(5),
			Channel:            ygot.Uint8(8),
			ChannelWidth:       ygot.Uint8(10),
			Scanning:           ygot.Bool(true),
			ScanningInterval:   ygot.Uint8(30),
		},
	}
}

func addWLANs(ap *ocstruct.WifiOffice_OfficeAp, addAuthWLAN bool) {
	wlans := &ocstruct.WifiOffice_OfficeAp_Ssids{}
	wlans.Ssid = make(map[string]*ocstruct.WifiOffice_OfficeAp_Ssids_Ssid)
	ap.Ssids = wlans

	wlans.Ssid[guestWLANName] = &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid{
		Name: ygot.String(guestWLANName),
		Config: &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config{
			AdvertiseApname:    ygot.Bool(false),
			BasicDataRates:     []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
			BroadcastFilter:    ygot.Bool(false),
			Csa:                ygot.Bool(false),
			DhcpRequired:       ygot.Bool(true),
			Dot11K:             ygot.Bool(false),
			Dva:                ygot.Bool(false),
			Enabled:            ygot.Bool(true),
			GtkTimeout:         ygot.Uint16(1000),
			Hidden:             ygot.Bool(false),
			MulticastFilter:    ygot.Bool(false),
			Name:               ygot.String(guestWLANName),
			OperatingFrequency: ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2_5_GHZ,
			Opmode:             ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config_Opmode_OPEN,
			PtkTimeout:         ygot.Uint16(1000),
			SupportedDataRates: []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
			VlanId:             ygot.Uint16(666),
		},
	}

	// Add auth WLAN.
	if !addAuthWLAN {
		return
	}

	wlans.Ssid[authWLANName] = &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid{
		Name: ygot.String(authWLANName),
		Config: &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config{
			AdvertiseApname:    ygot.Bool(false),
			BasicDataRates:     []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
			BroadcastFilter:    ygot.Bool(false),
			Csa:                ygot.Bool(false),
			DhcpRequired:       ygot.Bool(true),
			Dot11K:             ygot.Bool(false),
			Dva:                ygot.Bool(false),
			Enabled:            ygot.Bool(true),
			GtkTimeout:         ygot.Uint16(1000),
			Hidden:             ygot.Bool(false),
			MulticastFilter:    ygot.Bool(false),
			Name:               ygot.String(authWLANName),
			OperatingFrequency: ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2_5_GHZ,
			Opmode:             ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config_Opmode_OPEN,
			PtkTimeout:         ygot.Uint16(1000),
			SupportedDataRates: []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
			VlanId:             ygot.Uint16(250),
		},
	}
}

func generateConfig(apNum int, addAuthWLAN bool) *ocstruct.Office {
	// Generate a WIFI configuration for an office.
	office := &ocstruct.Office{}

	office.OfficeName = ygot.String(officeName)
	apVendor, err := office.NewVendor(vendor)
	if err != nil {
		log.Exitf("Error setting vendor: %v", err)
	}
	apVendor.VendorName = ygot.String(vendor)

	// Set up AP (vendor neutral).
	addAPs(office, apNum, addAuthWLAN)

	return office
}

func main() {
	office := generateConfig(1, true)

	jsonString, err := ygot.EmitJSON(office, &ygot.EmitJSONConfig{
		Format: ygot.RFC7951,
		Indent: "  ",
		RFC7951Config: &ygot.RFC7951JSONConfig{
			AppendModuleName: false,
		},
	})
	if err != nil {
		log.Exitf("Error outputting the configuration to JSON: %v", err)
	}
	log.Infof("Original office configJSON output:\n%v\n", jsonString)

	loadedOffice := &ocstruct.Office{}
	err = ocstruct.Unmarshal([]byte(jsonString), loadedOffice)
	if err != nil {
		log.Exitf("Error unmarshal JSON: %v", err)
	}

	loadedJSONString, err := ygot.EmitJSON(loadedOffice, &ygot.EmitJSONConfig{
		Format: ygot.RFC7951,
		Indent: "  ",
		RFC7951Config: &ygot.RFC7951JSONConfig{
			AppendModuleName: false,
		},
	})
	if err != nil {
		log.Exitf("Error outputting the loaded configuration to JSON: %v", err)
	}
	log.Infof("Loaded config JSON output:\n%v\n", loadedJSONString)
}
