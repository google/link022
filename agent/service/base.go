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
	"fmt"

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
func ApplyConfig(officeConfig *ocstruct.Office, setupIntf bool, deviceHostname, ethIntfName, wlanINTFName string) error {
	officeAPs := officeConfig.OfficeAp
	if _, ok := officeAPs[deviceHostname]; !ok {
		return fmt.Errorf("not found the configuration for AP %v", deviceHostname)
	}

	officeAP := officeAPs[deviceHostname]
	log.Infof("Configuring AP %v...", deviceHostname)

	if setupIntf {
		// Configure eth interface.
		if err := configEthIntf(ethIntfName, ocutil.VLANIDs(officeConfig)); err != nil {
			return err
		}

		//Configure WLAN interface.
		if err := configWLANIntf(wlanINTFName); err != nil {
			return err
		}
	}

	// Configure hostapd.
	if err := configHostapd(officeAP, wlanINTFName); err != nil {
		return err
	}

	return nil
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
