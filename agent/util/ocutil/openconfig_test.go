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

package ocutil

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/link022/agent/util/mock"
	"github.com/google/link022/generated/ocstruct"
)

func TestVLANIDs(t *testing.T) {
	// Define test cases.
	tests := []struct {
		apConfig *ocstruct.Device
		vlanIDs  []int
	}{{
		apConfig: mock.GenerateConfig(true),
		vlanIDs:  []int{250, 666},
	}, {
		apConfig: mock.GenerateConfig(false),
		vlanIDs:  []int{666},
	}, {
		apConfig: nil,
		vlanIDs:  []int{},
	}}

	for _, test := range tests {
		got := VLANIDs(test.apConfig)
		want := test.vlanIDs
		sort.Ints(got)
		sort.Ints(want)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Incorrect VLAN IDs (got: %v, want: %v).", got, want)
		}
	}
}

func TestVLANChanged(t *testing.T) {
	// Define test cases.
	tests := []struct {
		existingConfig      *ocstruct.Device
		updatedConfigConfig *ocstruct.Device
		vlanChanged         bool
	}{{
		existingConfig:      mock.GenerateConfig(true),
		updatedConfigConfig: mock.GenerateConfig(true),
		vlanChanged:         false,
	}, {
		existingConfig:      mock.GenerateConfig(true),
		updatedConfigConfig: mock.GenerateConfig(false),
		vlanChanged:         true,
	}, {
		existingConfig:      mock.GenerateConfig(false),
		updatedConfigConfig: mock.GenerateConfig(true),
		vlanChanged:         true,
	}, {
		existingConfig:      nil,
		updatedConfigConfig: mock.GenerateConfig(true),
		vlanChanged:         true,
	}, {
		existingConfig:      mock.GenerateConfig(true),
		updatedConfigConfig: nil,
		vlanChanged:         true,
	}, {
		existingConfig:      nil,
		updatedConfigConfig: nil,
		vlanChanged:         false,
	}}

	for _, test := range tests {
		got := VLANChanged(VLANIDs(test.existingConfig), VLANIDs(test.updatedConfigConfig))
		want := test.vlanChanged
		if got != want {
			t.Errorf("Incorrect result (got: %v, want: %v).", got, want)
		}
	}
}

func TestRadiusServers(t *testing.T) {
	// Define test cases.
	tests := []struct {
		apConfig      *ocstruct.Device
		radiusServers map[string]*ocstruct.OpenconfigOfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server
	}{{
		apConfig:      mock.GenerateConfig(false),
		radiusServers: make(map[string]*ocstruct.OpenconfigOfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server),
	}, {
		apConfig: mock.GenerateConfig(true),
		radiusServers: map[string]*ocstruct.OpenconfigOfficeAp_System_Aaa_ServerGroups_ServerGroup_Servers_Server{
			mock.AuthWLANName: mock.RadiusServer(),
		},
	}}

	for _, test := range tests {
		got := RadiusServers(test.apConfig)
		want := test.radiusServers
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Incorrect result (got: %v, want: %v).", got, want)
		}
	}
}
