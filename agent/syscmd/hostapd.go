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

// StartHostapd starts a hostapd process link to the given WLAN interface.
func (r *CommandRunner) StartHostapd(configFilePath string) error {
	log.Infof("Starting hostapd process with config file: %v...", configFilePath)
	if _, err := r.ExecCommand(false, "hostapd", configFilePath); err != nil {
		return err
	}
	log.Infof("Started a hostapd with config file: %v.", configFilePath)
	return nil
}

// StopAllHostapd kills all running hostapd processes.
func (r *CommandRunner) StopAllHostapd() error {
	log.Info("Stopping all hostapd processes...")
	if _, err := r.ExecCommand(true, "killall", "-q", "hostapd"); err != nil {
		return err
	}
	log.Info("Stopped all hostapd processes.")
	return nil
}
