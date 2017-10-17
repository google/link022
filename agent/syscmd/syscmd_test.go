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
	"fmt"
	"reflect"
	"sort"
	"testing"
)

const (
	testIPLinkInfo = `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1\    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00 promiscuity 0 
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP mode DEFAULT group default qlen 1000\    link/ether b8:27:eb:ef:4e:b6 brd ff:ff:ff:ff:ff:ff promiscuity 2 
3: wlan0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc pfifo_fast state DOWN mode DEFAULT group default qlen 1000\    link/ether b8:27:eb:ba:1b:e3 brd ff:ff:ff:ff:ff:ff promiscuity 0 
4: wlan1: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc mq state DOWN mode DEFAULT group default qlen 1000\    link/ether 02:c0:ca:90:2f:50 brd ff:ff:ff:ff:ff:ff promiscuity 0 
5: eth0.250@eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master br_250 state UP mode DEFAULT group default qlen 1000\    link/ether b8:27:eb:ef:4e:b6 brd ff:ff:ff:ff:ff:ff promiscuity 1 \    vlan protocol 802.1Q id 250 <REORDER_HDR> \    bridge_slave 
6: eth0.666@eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue master br_666 state UP mode DEFAULT group default qlen 1000\    link/ether b8:27:eb:ef:4e:b6 brd ff:ff:ff:ff:ff:ff promiscuity 1 \    vlan protocol 802.1Q id 666 <REORDER_HDR> \    bridge_slave 
7: br_250: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000\    link/ether b8:27:eb:ef:4e:b6 brd ff:ff:ff:ff:ff:ff promiscuity 0 \    bridge 
8: br_666: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000\    link/ether b8:27:eb:ef:4e:b6 brd ff:ff:ff:ff:ff:ff promiscuity 0 \    bridge 
`

	testIntf     = "eth0"
	testWLANIntf = "wlan0"
	testVLANID   = 10

	bridgeName = "br_0"
)

var (
	testIPVLANLink = []int{250, 666}

	runner = &CommandRunner{
		ExecCommand: func(wait bool, command string, args ...string) (string, error) {
			if command == "ip" && reflect.DeepEqual(args, []string{"-o", "-d", "link", "show"}) {
				return testIPLinkInfo, nil
			}

			// No ops.
			return "", nil
		},
	}
)

// Testing Interface commands.

func TestCreateVLAN(t *testing.T) {
	vlanIntfName, err := runner.CreateVLAN(testIntf, testVLANID)
	if err != nil {
		t.Errorf("Creating VLAN interface failed. Error: %v.", err)
		return
	}

	expectedVLANIntfName := fmt.Sprintf("%s.%d", testIntf, testVLANID)
	if vlanIntfName != expectedVLANIntfName {
		t.Errorf("Incorrect VLAN interface name. (actual: %v, expected: %v)",
			vlanIntfName, expectedVLANIntfName)
	}
}

func TestDeleteVLAN(t *testing.T) {
	if err := runner.DeleteVLAN(testIntf, testVLANID); err != nil {
		t.Errorf("Deleting VLAN interface failed. Error: %v.", err)
	}
}

func TestRestartIntf(t *testing.T) {
	if err := runner.RestartIntf(testIntf); err != nil {
		t.Errorf("Restarting interface failed. Error: %v.", err)
	}
}

func TestWipeOutIntfIP(t *testing.T) {
	if err := runner.WipeOutIntfIP(testIntf); err != nil {
		t.Errorf("Wiping out interface IP failed. Error: %v.", err)
	}
}

func TestVLANOnIntf(t *testing.T) {
	if vlanIDs, err := runner.VLANOnIntf(testIntf); err != nil {
		t.Errorf("Fetching VLAN interface failed. Error: %v.", err)
	} else {
		sort.Ints(vlanIDs)
		if !reflect.DeepEqual(vlanIDs, testIPVLANLink) {
			t.Errorf("Incorrect result of VLANOnIntf, actual: %v, expected: %v.", vlanIDs, testIPVLANLink)
		}
	}
}

// Testing bridging commands.

func TestCreateBridge(t *testing.T) {
	if err := runner.CreateBridge(bridgeName); err != nil {
		t.Errorf("Creating bridge failed. Error: %v.", err)
	}
}

func TestDeleteBridge(t *testing.T) {
	if err := runner.DeleteBridge(bridgeName); err != nil {
		t.Errorf("Deleting bridge failed. Error: %v.", err)
	}
}

func TestAddBridgeIntf(t *testing.T) {
	if err := runner.AddBridgeIntf(bridgeName, testIntf); err != nil {
		t.Errorf("Adding bridge interface failed. Error: %v.", err)
	}
}

// Test hostapd commands.

func TestStartHostapd(t *testing.T) {
	if err := runner.StartHostapd(testWLANIntf); err != nil {
		t.Errorf("Starting hostapd process failed. Error: %v.", err)
	}
}

func TestStopAllHostapd(t *testing.T) {
	if err := runner.StopAllHostapd(); err != nil {
		t.Errorf("Stopping hostapd processes failed. Error: %v.", err)
	}
}
