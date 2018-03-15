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

// Package service contains methods that manage Link022 AP device.
package service

import (
	log "github.com/golang/glog"
	"github.com/google/link022/agent/syscmd"
	"github.com/google/link022/agent/util/ocutil"
	"github.com/google/link022/generated/ocstruct"
)

var (
	cmdRunner = syscmd.Runner()
	runFolder = "/var/run/link022"
)

// ApplyConfig configures this device to a Link022 AP based on the given configuration.
func ApplyConfig(officeAP *ocstruct.OpenconfigAccessPoints_AccessPoints_AccessPoint, setupIntf bool, ethIntfName, wlanINTFName string) error {
	log.Infof("Configuring AP %s...", *officeAP.Hostname)

	if setupIntf {
		// Configure eth interface.
		if err := configEthIntf(ethIntfName, ocutil.VLANIDs(officeAP)); err != nil {
			return err
		}

		//Configure WLAN interface.
		if err := configWLANIntf(wlanINTFName); err != nil {
			return err
		}
	}

	// Configure hostapd.
	return configHostapd(officeAP, wlanINTFName)
}

// CleanupConfig cleans up the current AP configuration on this device.
// It goes through all cleanup steps even if some failures are detected, and returns all errors.
func CleanupConfig(ethIntfName string, vlanIDs []int) []error {
	var errs []error

	// Stop hostapd processes.
	if err := cmdRunner.StopAllHostapd(); err != nil {
		errs = append(errs, err)
	}

	// Clean up eth interfaces.
	if len(vlanIDs) > 0 {
		errs = append(errs, cleanupEthIntf(ethIntfName, vlanIDs)...)
	}

	log.Infof("Cleaned up AP. Number of errors = %d.", len(errs))
	return errs
}
