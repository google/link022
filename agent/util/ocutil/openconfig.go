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

// Package ocutil contains helper functions related to OpenConfig models.
package ocutil

import (
	"reflect"
	"sort"

	"github.com/google/link022/generated/ocstruct"
)

// FindAPConfig finds the configuration of the AP with a specific hostname.
// It returns nil if not matching AP found.
func FindAPConfig(apConfigs *ocstruct.Device, hostname string) *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint {
	if apConfigs.AccessPoints == nil {
		return nil
	}

	apConfig, ok := apConfigs.AccessPoints.AccessPoint[hostname]
	if !ok {
		return nil
	}
	return apConfig
}

// VLANChanged checkes whether there is any difference between the given two VLAN ID lists.
func VLANChanged(existingVLANIDs, updatedVLANIDs []int) bool {
	sort.Ints(existingVLANIDs)
	sort.Ints(updatedVLANIDs)
	return !reflect.DeepEqual(existingVLANIDs, updatedVLANIDs)
}

// VLANIDs fetches the ID of all VLANs appears in the given office configuration.
func VLANIDs(apConfig *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint) []int {
	vlanIDs := []int{}

	if apConfig == nil {
		return vlanIDs
	}

	wlans := apConfig.Ssids
	if wlans == nil || len(wlans.Ssid) == 0 {
		return vlanIDs
	}

	for _, wlan := range wlans.Ssid {
		vlanIDs = append(vlanIDs, int(*wlan.Config.DefaultVlan))
	}

	return vlanIDs
}

// RadiusServers fetches the radius server assigned to the given AP.
// It returns a SSID -> RadiusServer map.
func RadiusServers(ap *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint) map[string]*ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint_System_Aaa_ServerGroups_ServerGroup_Servers_Server {
	wlanRadiusMap := make(map[string]*ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint_System_Aaa_ServerGroups_ServerGroup_Servers_Server)
	if ap.Ssids == nil {
		return wlanRadiusMap
	}

	apServerGPs := aaaServerGroups(ap)
	if len(apServerGPs) == 0 {
		return wlanRadiusMap
	}

	for wlanName, wlan := range ap.Ssids.Ssid {
		if wlan.Config.ServerGroup == nil {
			continue
		}

		aaaServerGPName := *wlan.Config.ServerGroup
		if serverGP, ok := apServerGPs[aaaServerGPName]; ok {
			if radiusServer := aaaRadiusServer(serverGP); radiusServer != nil {
				wlanRadiusMap[wlanName] = radiusServer
			}
		}
	}

	return wlanRadiusMap
}

func aaaServerGroups(ap *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint) map[string]*ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint_System_Aaa_ServerGroups_ServerGroup {
	apSystemInfo := ap.System
	if apSystemInfo == nil {
		return nil
	}

	aaaInfo := apSystemInfo.Aaa
	if aaaInfo == nil {
		return nil
	}

	serverGP := aaaInfo.ServerGroups
	if serverGP == nil {
		return nil
	}

	return serverGP.ServerGroup
}

func aaaRadiusServer(serverGP *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint_System_Aaa_ServerGroups_ServerGroup) *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint_System_Aaa_ServerGroups_ServerGroup_Servers_Server {
	if serverGP == nil || serverGP.Config == nil || serverGP.Config.Type != ocstruct.OpenconfigAaaTypes_AAA_SERVER_TYPE_RADIUS {
		return nil
	}

	if serverGP.Servers == nil {
		return nil
	}

	for _, radiusServer := range serverGP.Servers.Server {
		if radiusServer != nil {
			// Return the first radius server specified.
			return radiusServer
		}
	}
	// Not found Radius server
	return nil
}
