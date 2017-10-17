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

package syscmd

import (
	log "github.com/golang/glog"
)

// CreateBridge creates a network bridge with a certain name.
func (r *CommandRunner) CreateBridge(bridgeName string) error {
	log.Infof("Creating bridge %v...", bridgeName)
	if _, err := r.ExecCommand(true, "brctl", "addbr", bridgeName); err != nil {
		return err
	}
	log.Infof("Created bridge %v.", bridgeName)
	return nil
}

// DeleteBridge deletes a network bridge with a certain name.
func (r *CommandRunner) DeleteBridge(bridgeName string) error {
	log.Infof("Deleting bridge %v...", bridgeName)
	if _, err := r.ExecCommand(true, "brctl", "delbr", bridgeName); err != nil {
		return err
	}
	log.Infof("Deleted bridge %v.", bridgeName)
	return nil
}

// AddBridgeIntf adds an interface to a network bridge.
func (r *CommandRunner) AddBridgeIntf(bridgeName, intfName string) error {
	log.Infof("Adding interface %v to bridge %v...", intfName, bridgeName)
	if _, err := r.ExecCommand(true, "brctl", "addif", bridgeName, intfName); err != nil {
		return err
	}
	log.Infof("Added interface %v to bridge %v.", intfName, bridgeName)
	return nil
}
