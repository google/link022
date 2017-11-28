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

package mock

import (
	"github.com/google/link022/generated/ocstruct"
	"github.com/openconfig/ygot/ygot"
)

var (
	officeName = "test-office"
	vendorName = "link022"

	apNames = [2]string{"test-pi-1", "test-pi-2"}

	radiusServerGroupName = "radius-server-group"
	radiusServerAddr      = "1.1.1.1"

	GuestWLANName = "Guest-Emu"
	AuthWLANName  = "Auth-Emu"
)

// GenerateConfig generates a office wireless configuration for test.
func GenerateConfig(apNum int, addAuthWLAN bool) *ocstruct.Office {
	// Generate a WIFI configuration for an OWCA office.
	office := &ocstruct.Office{}

	office.OfficeName = ygot.String(officeName)
	office.Vendor = make(map[string]*ocstruct.WifiOffice_Vendor)
	office.Vendor[vendorName] = &ocstruct.WifiOffice_Vendor{
		VendorName: ygot.String(vendorName),
	}

	// Set up OWCA-AP (vendor neutral).
	office.OfficeAp = make(map[string]*ocstruct.WifiOffice_OfficeAp)
	for i := 0; i < apNum; i++ {
		if i > 1 {
			break
		}

		apName := apNames[i]
		office.OfficeAp[apName] = apConfig(apName, addAuthWLAN)
	}

	return office
}

// Test configuration generator.
func apConfig(apName string, addAuthWLAN bool) *ocstruct.WifiOffice_OfficeAp {
	ap := &ocstruct.WifiOffice_OfficeAp{
		Hostname: ygot.String(apName),
		Vendor:   ygot.String(vendorName),
	}

	if addAuthWLAN {
		ap.System = systemInfo()
	}

	ap.Radios = radios()
	ap.Ssids = wlans(addAuthWLAN)
	return ap
}

func RadiusServer() *ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server {
	return &ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server{
		Address: ygot.String(radiusServerAddr),
		Radius: &ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server_Radius{
			Config: &ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server_Radius_Config{
				AuthPort:  ygot.Uint16(1812),
				SecretKey: ygot.String("radiuspwd"),
			},
		},
	}
}

func systemInfo() *ocstruct.WifiOffice_OfficeAp_System {
	return &ocstruct.WifiOffice_OfficeAp_System{
		Aaa: &ocstruct.WifiOffice_OfficeAp_System_Aaa{
			ServerGroups: &ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups{
				ServerGroup: map[string]*ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups_ServerGroup{
					radiusServerGroupName: {
						Name: ygot.String(radiusServerGroupName),
						Config: &ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups_ServerGroup_Config{
							Type: ocstruct.OpenconfigAaaTypes_AAA_SERVER_TYPE_RADIUS,
						},
						Servers: &ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers{
							Server: map[string]*ocstruct.WifiOffice_OfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server{
								radiusServerAddr: RadiusServer(),
							},
						},
					},
				},
			},
		},
	}
}

func radios() *ocstruct.WifiOffice_OfficeAp_Radios {
	radios := &ocstruct.WifiOffice_OfficeAp_Radios{}
	radios.Radio = make(map[uint8]*ocstruct.WifiOffice_OfficeAp_Radios_Radio)

	radioID := uint8(1)
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

	return radios
}

func wlans(addAuthWLAN bool) *ocstruct.WifiOffice_OfficeAp_Ssids {
	wlans := &ocstruct.WifiOffice_OfficeAp_Ssids{}
	wlans.Ssid = make(map[string]*ocstruct.WifiOffice_OfficeAp_Ssids_Ssid)

	wlans.Ssid[GuestWLANName] = guestWLAN()

	// Add auth WLAN.
	if addAuthWLAN {
		wlans.Ssid[AuthWLANName] = authWLAN()
	}

	return wlans
}

func guestWLAN() *ocstruct.WifiOffice_OfficeAp_Ssids_Ssid {
	return &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid{
		Name: ygot.String(GuestWLANName),
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
			Name:               ygot.String(GuestWLANName),
			OperatingFrequency: ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2_5_GHZ,
			Opmode:             ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config_Opmode_OPEN,
			PtkTimeout:         ygot.Uint16(1000),
			SupportedDataRates: []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
			VlanId:             ygot.Uint16(666),
		},
	}
}

func authWLAN() *ocstruct.WifiOffice_OfficeAp_Ssids_Ssid {
	return &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid{
		Name: ygot.String(AuthWLANName),
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
			Name:               ygot.String(AuthWLANName),
			OperatingFrequency: ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2_5_GHZ,
			Opmode:             ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config_Opmode_WPA2_ENTERPRISE,
			ServerGroup:        ygot.String(radiusServerGroupName),
			PtkTimeout:         ygot.Uint16(1000),
			SupportedDataRates: []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
			VlanId:             ygot.Uint16(250),
		},
	}
}
