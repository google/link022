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

// Package gnmi contains GNMI server and related methods.
package gnmi

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"

	"github.com/google/gnxi/gnmi"
	"github.com/google/link022/generated/ocstruct"

	log "github.com/golang/glog"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	runFolder        = "/var/run/link022"
	apConfigFileName = "link022.conf"
)

var (
	// link022ModelData is a list of models supported in this GNMI server.
	link022ModelData = []*pb.ModelData{{
		Name:         "openconfig-access-points",
		Organization: "OpenConfig working group",
		Version:      "0.1.0",
	}}
)

// Server is a GNMI server.
type Server struct {
	*gnmi.Server
}

// NewServer creates a GNMI server.
func NewServer() (*Server, error) {
	// Load existing config
	initConfigContent, err := loadExistingConfigContent()
	if err != nil {
		log.Errorf("Failed to load the existing configuration. Error: %v.", err)
		initConfigContent = nil
	}

	// Create the GNMI server.
	model := gnmi.NewModel(link022ModelData,
		reflect.TypeOf((*ocstruct.Device)(nil)),
		ocstruct.SchemaTree["Device"],
		ocstruct.Unmarshal,
		ocstruct.ΛEnum)

	s, err := gnmi.NewServer(model,
		initConfigContent,
		handleSet)
	if err != nil {
		return nil, err
	}

	gnmiServer := &Server{s}
	log.Info("GNMI server created.")
	return gnmiServer, nil
}

func loadExistingConfigContent() ([]byte, error) {
	existingConfigFilePath := path.Join(runFolder, apConfigFileName)

	if _, err := os.Stat(existingConfigFilePath); os.IsNotExist(err) {
		log.Info("No existing configuration found.")
		return nil, nil
	}

	existingConfigContent, err := ioutil.ReadFile(existingConfigFilePath)
	if err != nil {
		return nil, err
	}

	log.Info("Loaded existing configuration.")
	return existingConfigContent, nil
}
