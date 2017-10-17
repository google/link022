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

func VLANChanged(existingVLANIDs, updatedVLANIDs []int) bool {
	sort.Ints(existingVLANIDs)
	sort.Ints(updatedVLANIDs)
	return !reflect.DeepEqual(existingVLANIDs, updatedVLANIDs)
}

// VLANIDs fetches the ID of all VLANs appears in the given office configuration.
func VLANIDs(officeConfig *ocstruct.Office) []int {
	vlanIDs := []int{}

	if officeConfig == nil {
		return vlanIDs
	}

	for _, ap := range officeConfig.OfficeAp {
		wlans := ap.Ssids
		if wlans == nil || len(wlans.Ssid) == 0 {
			continue
		}

		for _, wlan := range wlans.Ssid {
			vlanIDs = append(vlanIDs, int(*wlan.Config.VlanId))
		}
	}

	return vlanIDs
}
