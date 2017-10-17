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
		officeConfig *ocstruct.Office
		vlanIDs      []int
	}{{
		officeConfig: mock.GenerateConfig(1, true),
		vlanIDs:      []int{250, 666},
	}, {
		officeConfig: mock.GenerateConfig(1, false),
		vlanIDs:      []int{666},
	}, {
		officeConfig: mock.GenerateConfig(0, false),
		vlanIDs:      []int{},
	}}

	for _, test := range tests {
		got := VLANIDs(test.officeConfig)
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
		existingConfig      *ocstruct.Office
		updatedConfigConfig *ocstruct.Office
		vlanChanged         bool
	}{{
		existingConfig:      mock.GenerateConfig(1, true),
		updatedConfigConfig: mock.GenerateConfig(1, true),
		vlanChanged:         false,
	}, {
		existingConfig:      mock.GenerateConfig(1, true),
		updatedConfigConfig: mock.GenerateConfig(1, false),
		vlanChanged:         true,
	}, {
		existingConfig:      mock.GenerateConfig(1, false),
		updatedConfigConfig: mock.GenerateConfig(1, true),
		vlanChanged:         true,
	}, {
		existingConfig:      mock.GenerateConfig(0, true),
		updatedConfigConfig: mock.GenerateConfig(1, true),
		vlanChanged:         true,
	}, {
		existingConfig:      mock.GenerateConfig(1, true),
		updatedConfigConfig: mock.GenerateConfig(0, true),
		vlanChanged:         true,
	}, {
		existingConfig:      mock.GenerateConfig(0, true),
		updatedConfigConfig: mock.GenerateConfig(0, true),
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
