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
	"testing"

	"github.com/google/link022/generated/ocstruct"
	"github.com/openconfig/ygot/ygot"
)

// TestYGOT verifies the current version of ygot library works with Link022 agent.
func TestYGOT(t *testing.T) {
	verifyOfficeConfig(t, GenerateConfig(true))
	verifyOfficeConfig(t, GenerateConfig(false))
	verifyOfficeConfig(t, GenerateConfig(true))
	verifyOfficeConfig(t, GenerateConfig(false))
	verifyOfficeConfig(t, GenerateConfig(true))
	verifyOfficeConfig(t, GenerateConfig(false))
}

func verifyOfficeConfig(t *testing.T, apConfig *ocstruct.Device) {
	// Test validation.
	if err := apConfig.Validate(); err != nil {
		t.Errorf("AP configuration is not valid. Error: %v.", err)
	}

	// Test marshalling.
	jsonString, err := ygot.EmitJSON(apConfig, &ygot.EmitJSONConfig{
		Format: ygot.RFC7951,
		Indent: "  ",
		RFC7951Config: &ygot.RFC7951JSONConfig{
			AppendModuleName: true,
		},
	})
	if err != nil {
		t.Errorf("Marshalling the AP configuration to JSON failed. Error: %v.", err)
	}

	// Test unmarshalling.
	loadedConfig := &ocstruct.Device{}
	if err = ocstruct.Unmarshal([]byte(jsonString), loadedConfig); err != nil {
		t.Errorf("Unmarshalling AP configuration JSON string to OpenConfig structs failed. Error: %v.", err)
	}

	// Test validation.
	if err := loadedConfig.Validate(); err != nil {
		t.Errorf("The loaded AP configuration is not valid. Error: %v.", err)
	}

	// Test marshalling again.
	loadedJSONString, err := ygot.EmitJSON(loadedConfig, &ygot.EmitJSONConfig{
		Format: ygot.RFC7951,
		Indent: "  ",
		RFC7951Config: &ygot.RFC7951JSONConfig{
			AppendModuleName: true,
		},
	})
	if err != nil {
		t.Errorf("Marshalling the loaded AP configuration to JSON failed. Error: %v.", err)
	}

	if jsonString != loadedJSONString {
		t.Errorf("The loaded AP configuration does not match the original one, got: %s, expected: %s.", jsonString, loadedJSONString)
	}
}
