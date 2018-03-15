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

func TestFindAPConfig(t *testing.T) {
	// Define test cases.
	tests := []struct {
		apConfigs    *ocstruct.Device
		targetAPName string
		found        bool
	}{{
		apConfigs:    mock.GenerateConfig(true),
		targetAPName: "fake AP",
		found:        false,
	}, {
		apConfigs:    mock.GenerateConfig(true),
		targetAPName: "test-pi-1",
		found:        true,
	}}

	for _, test := range tests {
		matchedConfig := FindAPConfig(test.apConfigs, test.targetAPName)
		foundMatch := matchedConfig != nil

		if foundMatch != test.found {
			t.Errorf("Incorrect FindAPConfig result (got: %v, want:%v).", foundMatch, test.found)
		}
	}
}

func TestVLANIDs(t *testing.T) {
	// Define test cases.
	tests := []struct {
		apConfig *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint
		vlanIDs  []int
	}{{
		apConfig: mock.GenerateAPConfig(true),
		vlanIDs:  []int{250, 666},
	}, {
		apConfig: mock.GenerateAPConfig(false),
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
		existingConfig      *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint
		updatedConfigConfig *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint
		vlanChanged         bool
	}{{
		existingConfig:      mock.GenerateAPConfig(true),
		updatedConfigConfig: mock.GenerateAPConfig(true),
		vlanChanged:         false,
	}, {
		existingConfig:      mock.GenerateAPConfig(true),
		updatedConfigConfig: mock.GenerateAPConfig(false),
		vlanChanged:         true,
	}, {
		existingConfig:      mock.GenerateAPConfig(false),
		updatedConfigConfig: mock.GenerateAPConfig(true),
		vlanChanged:         true,
	}, {
		existingConfig:      nil,
		updatedConfigConfig: mock.GenerateAPConfig(true),
		vlanChanged:         true,
	}, {
		existingConfig:      mock.GenerateAPConfig(true),
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
		apConfig      *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint
		radiusServers map[string]*ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint_System_Aaa_ServerGroups_ServerGroup_Servers_Server
	}{{
		apConfig:      mock.GenerateAPConfig(false),
		radiusServers: make(map[string]*ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint_System_Aaa_ServerGroups_ServerGroup_Servers_Server),
	}, {
		apConfig: mock.GenerateAPConfig(true),
		radiusServers: map[string]*ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint_System_Aaa_ServerGroups_ServerGroup_Servers_Server{
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
