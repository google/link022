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

package service

import (
    "fmt"
    "io/ioutil"
    "path"
    "reflect"
    "testing"

    "github.com/google/link022/agent/syscmd"
    "github.com/google/link022/agent/util/mock"
    "github.com/google/link022/agent/util/ocutil"
    "github.com/google/link022/generated/ocstruct"
)

var (
    testHostname = "test-pi-1"
    testETHIntf = "eth0"
    testWLANIntf = "wlan0"
    testWLANIntfOriginMAC = "aa:bb:cc:dd:ee:ff"
    testWLANIntfUpdatedMAC = "02:bb:cc:dd:ee:f0"

    testSystemState *systemState
)

// Mock environment.

// systemState contains the state of mocked environment.
type systemState struct {
    Intfs map[string]bool  // interface name -> is up
    IntfMACs map[string]string // interface name -> MAC address
    NetworkBRs map[string][]string  // bridge name -> linked interfaces
    Hostapds map[string]bool // hostapd config file path -> started
}

type commandError struct {
    exitCode int
}

func (e *commandError) Error() string {
    return fmt.Sprintf("command failed with exit code %d", e.exitCode)
}

// executeMockCommand runs mocks system calls.
func executeMockCommand(wait bool, cmd string, args ...string) (string, error) {
    switch cmd {
    case "ip":
        switch {
        // Add ip link
        case len(args) == 10 && args[0] == "link" && args[1] == "add" && args[2] == "link":
            intfName := args[3]
            vlanIntfName := args[5]
            linkType := args[7]

            if intfName != testETHIntf{
                return fmt.Sprintf("Interface %v not found.\n", intfName), &commandError{2}
            }
            if linkType != "vlan" {
                return fmt.Sprintf("Not support link type %v.\n", linkType), &commandError{2}
            }

            if _, ok := testSystemState.Intfs[vlanIntfName]; ok {
                return fmt.Sprintf("Interface %v already exists.\n", vlanIntfName), &commandError{2}
            }

            testSystemState.Intfs[vlanIntfName] = true
            return "", nil
        // Delete a link
        case len(args) == 3 && args[0] == "link" && args[1] == "delete":
            vlanIntfName := args[2]
            if _, ok := testSystemState.Intfs[vlanIntfName]; !ok {
                return fmt.Sprintf("Interface %v does not exist.\n", vlanIntfName), &commandError{2}
            }
            delete(testSystemState.Intfs, vlanIntfName)
            return "",nil
        }
    case "brctl":
        switch {
        case len(args) == 2 && args[0] == "addbr":
            // Add bridge
            brName := args[1]
            if _, ok := testSystemState.NetworkBRs[brName]; ok {
                return fmt.Sprintf("Bridge %v already exists.\n", brName), &commandError{2}
            }
            testSystemState.NetworkBRs[brName] = []string{}
            testSystemState.Intfs[brName] = true
            return "", nil
        case len(args) == 2 && args[0] == "delbr":
            // Delete bridge
            brName := args[1]
            if _, ok := testSystemState.NetworkBRs[brName]; !ok {
                return fmt.Sprintf("Bridge %v does not exist.\n", brName), &commandError{2}
            }
            if testSystemState.Intfs[brName] {
                return fmt.Sprintf("Bridge %v is still up.\n", brName), &commandError{2}
            }
            delete(testSystemState.Intfs, brName)
            delete(testSystemState.NetworkBRs, brName)
            return "", nil
        case len(args) > 2 && args[0] == "addif":
            // Add bridge interface
            brName := args[1]
            intfName := args[2]
            if _, ok := testSystemState.Intfs[intfName]; !ok {
                return fmt.Sprintf("Interface %v not found.\n", intfName), &commandError{2}
            }
            if intfs, ok := testSystemState.NetworkBRs[brName]; ok {
                testSystemState.NetworkBRs[brName] = append(intfs, intfName)
                return "", nil
            } else {
                return fmt.Sprintf("Bridge %v not found.\n", brName), &commandError{2}
            }
        }
    case "ifconfig":
        switch {
        case len(args) == 2:
            intfName := args[0]
            if _, ok := testSystemState.Intfs[intfName]; !ok {
                return fmt.Sprintf("Interface %v not found.\n", intfName), &commandError{2}
            }
            if args[1] == "up" {
                testSystemState.Intfs[intfName] = true
            } else if args[1] == "down" {
                testSystemState.Intfs[intfName] = false
            }
            return "", nil
        case len(args) == 4 && args[1] == "hw" && args[2] == "ether":
            intfName := args[0]
            if _, ok := testSystemState.Intfs[intfName]; !ok {
                return fmt.Sprintf("Interface %v not found.\n", intfName), &commandError{2}
            }
            testSystemState.IntfMACs[intfName] = args[3]
            return "", nil
        }
    case "hostapd":
        if len(args) == 1 {
            hostapdConfigFile := args[0]
            if _, ok := testSystemState.Hostapds[hostapdConfigFile]; ok {
                return fmt.Sprintf("Hostapd with config file %v already started.\n", hostapdConfigFile), &commandError{2}
            }
            testSystemState.Hostapds[hostapdConfigFile] = true
            return "", nil
        }
    case "cat":
        if len(args) == 1 && args[0] == fmt.Sprintf("/sys/class/net/%s/address", testWLANIntf) {
            return testWLANIntfOriginMAC+"\n", nil
        }
    case "udhcpc":
        return "", nil
    case "killall":
        if len(args) == 2 && args[1] == "hostapd" {
            // Stop all hostapd processes.
            testSystemState.Hostapds = map[string]bool{}
            return "", nil
        }
    default:
        return fmt.Sprintf("Unknown command %v\n", cmd), &commandError{2}
    }

    return fmt.Sprintf("Invalid %s command arguments: %v\n", cmd, args), &commandError{2}
}

func TestApplyConfig(t *testing.T) {
    // Define test cases.
    type testCase struct {
        officeConfig *ocstruct.Office
        expectedSystemState *systemState
        expectedError error
    }

    tempRunFolder, err := ioutil.TempDir("", "link022")
    if err != nil {
        t.Fatalf("Unable to create a temp run time folder. Skip all tests.")
    }
    testWLANHostapdConfigFile := path.Join(tempRunFolder,fmt.Sprintf("hostapd_%s.conf", testWLANIntf))
    testCases := map[string]*testCase {
        "TestConfigWithTwoWLANs": &testCase {
            officeConfig: mock.GenerateConfig(1, true),
            expectedSystemState: &systemState {
                Intfs: map[string]bool{
                    testETHIntf: true,
                    testWLANIntf: true,
                    "eth0.250": true,
                    "eth0.666": true,
                    "br_250": true,
                    "br_666": true,
                },
                IntfMACs: map[string]string{
                    testWLANIntf: testWLANIntfUpdatedMAC,
                },
                NetworkBRs: map[string][]string{
                    "br_250": []string{"eth0.250"},
                    "br_666": []string{"eth0.666"},
                },
                Hostapds: map[string]bool {
                    testWLANHostapdConfigFile: true,
                },
            },
            expectedError: nil,
        },
        "TestConfigWithOneWLAN": &testCase {
            officeConfig: mock.GenerateConfig(1, false),
            expectedSystemState: &systemState {
                Intfs: map[string]bool{
                    testETHIntf: true,
                    testWLANIntf: true,
                    "eth0.666": true,
                    "br_666": true,
                },
                IntfMACs: map[string]string{
                    testWLANIntf: testWLANIntfUpdatedMAC,
                },
                NetworkBRs: map[string][]string{
                    "br_666": []string{"eth0.666"},
                },
                Hostapds: map[string]bool {
                    testWLANHostapdConfigFile: true,
                },
            },
            expectedError: nil,
        },
    }

    // Start testing.
    cmdRunner = &syscmd.CommandRunner {ExecCommand: executeMockCommand,}
    originalRunFolder := runFolder
    runFolder = tempRunFolder
    defer func() {
        cmdRunner = syscmd.Runner()
        runFolder = originalRunFolder
    }()

    for testName, test := range testCases {
        // Clean up the test system state.
        testSystemState = cleanedSysteState()

        err := ApplyConfig(test.officeConfig, true, testHostname, testETHIntf, testWLANIntf)
        checkResult(t, testName, err, test.expectedError)
        checkResult(t, testName, testSystemState, test.expectedSystemState)
    }
}

func TestCleanupConfig(t *testing.T) {
    // Define test cases.
    tests := []struct {
        officeConfig *ocstruct.Office
        configRequired bool
        succeeded bool
    }{{
        officeConfig: mock.GenerateConfig(1, true),
        configRequired: true,
        succeeded: true,
    }, {
        officeConfig: mock.GenerateConfig(1, false),
        configRequired: true,
        succeeded: true,
    }, {
        officeConfig: mock.GenerateConfig(0, false),
        configRequired: false,
        succeeded: true,
    }, {
        officeConfig: mock.GenerateConfig(1, false),
        configRequired: false,
        succeeded: false,
    }}

    // Start testing.
    tempRunFolder, err := ioutil.TempDir("", "link022")
    if err != nil {
        t.Fatalf("Unable to create a temp run time folder. Skip all tests.")
    }
    cmdRunner = &syscmd.CommandRunner {ExecCommand: executeMockCommand,}
    originalRunFolder := runFolder
    runFolder = tempRunFolder
    defer func() {
        cmdRunner = syscmd.Runner()
        runFolder = originalRunFolder
    }()

    for i, test := range tests {
        // Clean up the test system state.
        testSystemState = cleanedSysteState()
        cleanedSystemState := cleanedSysteState()
        testName := fmt.Sprintf("TestCleanupConfig_%d", i)

        if test.configRequired {
            if err := ApplyConfig(test.officeConfig, true, testHostname, testETHIntf, testWLANIntf); err != nil {
                t.Errorf("[%s] Configuration failed. Error: %v.", testName, err)
            }
            // Clean up does not restore the MAC address.
            cleanedSystemState.IntfMACs[testWLANIntf] = testWLANIntfUpdatedMAC
        }

        errs := CleanupConfig(testETHIntf, ocutil.VLANIDs(test.officeConfig))
        checkResult(t, testName, len(errs) == 0, test.succeeded)
        checkResult(t, testName, testSystemState, cleanedSystemState)
    }
}

func cleanedSysteState() *systemState {
    return &systemState {
        Intfs: map[string]bool{
            testETHIntf: true,
            testWLANIntf: true,
        },
        IntfMACs: map[string]string{
            testWLANIntf: testWLANIntfOriginMAC,
        },
        NetworkBRs: map[string][]string{},
        Hostapds: map[string]bool{},
    }
}

func checkResult(t *testing.T, testName string, got, want interface{}) {
    if !reflect.DeepEqual(got, want) {
        t.Errorf("[%v] the test result is not correct (got: %v, want: %v)", testName, got, want)
    }
}

