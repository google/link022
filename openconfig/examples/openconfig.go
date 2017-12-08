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
	"flag"

	log "github.com/golang/glog"
	"github.com/google/link022/agent/util/mock"
	"github.com/google/link022/generated/ocstruct"
	"github.com/openconfig/ygot/ygot"
)

func main() {
	flag.Parse()
	apConfig := mock.GenerateConfig(true)

	jsonString, err := ygot.EmitJSON(apConfig, &ygot.EmitJSONConfig{
		Format: ygot.RFC7951,
		Indent: "  ",
		RFC7951Config: &ygot.RFC7951JSONConfig{
			AppendModuleName: false,
		},
	})
	if err != nil {
		log.Exitf("Error outputting the configuration to JSON: %v", err)
	}
	log.Infof("Original AP configJSON output:\n%v\n", jsonString)

	loadedAP := &ocstruct.Device{}
	err = ocstruct.Unmarshal([]byte(jsonString), loadedAP)
	if err != nil {
		log.Exitf("Error unmarshal JSON: %v", err)
	}

	loadedJSONString, err := ygot.EmitJSON(loadedAP, &ygot.EmitJSONConfig{
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
