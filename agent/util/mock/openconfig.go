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

    ap1Hostname = "test-pi-1"
    ap2Hostname = "test-pi-2"

    radioID = uint8(1)

    guestWLANName = "Guest-Emu"
    authWLANName  = "Auth-Emu"
)

// GenerateConfig generates a office wireless configuration for test.
func GenerateConfig(apNum int, addAuthWLAN bool) *ocstruct.Office {
    // Generate a WIFI configuration for an OWCA office.
    office := &ocstruct.Office{}

    office.OfficeName = ygot.String(officeName)
    office.Vendor = make(map[string]*ocstruct.WifiOffice_Vendor)
    office.Vendor[vendorName] = &ocstruct.WifiOffice_Vendor {
        VendorName: ygot.String(vendorName),
    }

    office.AuthServerConfig = &ocstruct.WifiOffice_AuthServerConfig {
        Name: ygot.String("1.1.1.1"),
        AuthPort: ygot.Uint16(1812),
        SecretKey: ygot.String("radiuspwd"),
    }

    // Set up OWCA-AP (vendor neutral).
    addAPs(office, apNum, addAuthWLAN)

    return office
}

// Test configuraton generator.

func addAPs(wifiOWCA *ocstruct.Office, apNum int, addAuthWLAN bool) {
    if apNum <= 0 {
        return
    }

    wifiOWCA.OfficeAp = make(map[string]*ocstruct.WifiOffice_OfficeAp)

    // Add AP 1.
    ap1 :=  &ocstruct.WifiOffice_OfficeAp {
        Hostname: ygot.String(ap1Hostname),
        Vendor: ygot.String(vendorName),
    }
    wifiOWCA.OfficeAp[ap1Hostname] = ap1
    addRadios(ap1)
    addWLANs(ap1, addAuthWLAN)

    if apNum <= 1 {
        return
    }

    // Add AP 2.
    ap2 :=  &ocstruct.WifiOffice_OfficeAp {
        Hostname: ygot.String(ap2Hostname),
        Vendor: ygot.String(vendorName),
    }
    wifiOWCA.OfficeAp[ap2Hostname] = ap2
    addRadios(ap2)
    addWLANs(ap2, addAuthWLAN)
}

func addRadios(ap *ocstruct.WifiOffice_OfficeAp) {
    radios := &ocstruct.WifiOffice_OfficeAp_Radios{}
    radios.Radio = make(map[uint8]*ocstruct.WifiOffice_OfficeAp_Radios_Radio)
    ap.Radios = radios

    radios.Radio[radioID] = &ocstruct.WifiOffice_OfficeAp_Radios_Radio{
        Id: ygot.Uint8(radioID),
        Config: &ocstruct.WifiOffice_OfficeAp_Radios_Radio_Config {
            Id: ygot.Uint8(radioID),
            Enabled: ygot.Bool(true),
            OperatingFrequency: ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2GHZ,
            TransmitPower: ygot.Uint8(5),
            Channel: ygot.Uint8(8),
            ChannelWidth: ygot.Uint8(10),
            Scanning: ygot.Bool(true),
            ScanningInterval: ygot.Uint8(30),
        },
    }
}

func addWLANs(ap *ocstruct.WifiOffice_OfficeAp, addAuthWLAN bool) {
    wlans := &ocstruct.WifiOffice_OfficeAp_Ssids{}
    wlans.Ssid = make(map[string]*ocstruct.WifiOffice_OfficeAp_Ssids_Ssid)
    ap.Ssids = wlans

    wlans.Ssid[guestWLANName] = &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid {
        Name: ygot.String(guestWLANName),
        Config: &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config {
            AdvertiseApname: ygot.Bool(false),
            BasicDataRates: []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
            BroadcastFilter: ygot.Bool(false),
            Csa: ygot.Bool(false),
            DhcpRequired: ygot.Bool(true),
            Dot11K: ygot.Bool(false),
            Dva: ygot.Bool(false),
            Enabled: ygot.Bool(true),
            GtkTimeout: ygot.Uint16(1000),
            Hidden: ygot.Bool(false),
            MulticastFilter: ygot.Bool(false),
            Name: ygot.String(guestWLANName),
            OperatingFrequency: ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2_5_GHZ,
            Opmode: ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config_Opmode_OPEN,
            PtkTimeout: ygot.Uint16(1000),
            SupportedDataRates: []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
            VlanId: ygot.Uint16(666),
        },
    }

    // Add auth WLAN.
    if !addAuthWLAN {
        return
    }

    wlans.Ssid[authWLANName] = &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid {
        Name: ygot.String(authWLANName),
        Config: &ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config {
            AdvertiseApname: ygot.Bool(false),
            BasicDataRates: []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
            BroadcastFilter: ygot.Bool(false),
            Csa: ygot.Bool(false),
            DhcpRequired: ygot.Bool(true),
            Dot11K: ygot.Bool(false),
            Dva: ygot.Bool(false),
            Enabled: ygot.Bool(true),
            GtkTimeout: ygot.Uint16(1000),
            Hidden: ygot.Bool(false),
            MulticastFilter: ygot.Bool(false),
            Name: ygot.String(authWLANName),
            OperatingFrequency: ocstruct.OpenconfigWifiTypes_OPERATING_FREQUENCY_FREQ_2_5_GHZ,
            Opmode: ocstruct.WifiOffice_OfficeAp_Ssids_Ssid_Config_Opmode_WPA2_ENTERPRISE,
            PtkTimeout: ygot.Uint16(1000),
            SupportedDataRates: []ocstruct.E_OpenconfigWifiTypes_DATA_RATE{ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_11MB, ocstruct.OpenconfigWifiTypes_DATA_RATE_RATE_24MB},
            VlanId: ygot.Uint16(250),
        },
    }
}
