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
	"errors"
	"net"
	"strings"

	log "github.com/golang/glog"
)

// DeviceIPv4 fetches the IPv4 address of the device.
func (r *CommandRunner) DeviceIPv4() (string, error) {
	ipInfo, err := r.ExecCommand(true, "hostname", "-I")
	if err != nil {
		return "", err
	}

	// Find the first IPv4 address.
	ipList := strings.Fields(ipInfo)
	for _, ipString := range ipList {
		ipAddr := net.ParseIP(ipString)
		if ipAddr != nil && ipAddr.To4() != nil {
			log.Infof("The device has IPv4 address %s.", ipString)
			return ipString, nil
		}
	}
	return "", errors.New("no IPv4 address found on this device")
}
