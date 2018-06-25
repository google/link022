/* Copyright 2018 Google Inc.

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

// Package common contains functions and models shared by all components.
package common

// OPType is the type of gNMI operation.
type OPType string

const (
	// OPReplace is the gNMI replace operation.
	OPReplace OPType = "replace"
	// OPUpdate is the gNMI update operation.
	OPUpdate OPType = "update"
	// OPDelete is the gNMI delete operation.
	OPDelete OPType = "delete"
	// OPGet is the gNMI get operation.
	OPGet OPType = "get"
	// OPSubscribe is the gNMI subscribe operation.
	OPSubscribe OPType = "subscribe"
)

// GNMITest is top-level model of gNMI test.
type GNMITest struct {
	// Name is the test name.
	Name string `json:"name"`
	// Description is the detail description of this test.
	Description string `json:"description"`
	// GNMITestCase is the list of test cases to run in this test.
	GNMITestCase []*TestCase `json:"test_cases"`
}

// TestCase describes a gNMI test cases.
type TestCase struct {
	// Name is the test case name.
	Name string `json:"name"`
	// Description is the detail description of this test case.
	Description string `json:"description"`
	// OPs contains a list of operations need to be processed in this test case.
	// All operations are processed in one single gNMI message.
	OPs []*Operation `json:"ops"`
}

// Operation represents a gNMI operation.
type Operation struct {
	// Type is the gNMI operation type.
	Type OPType `json:"type"`
	// Path is the xPath of the target field/branch.
	Path string `json:"path"`
	// Val is the string format of the desired value.
	// Supported types:
	//     Integer: "1", "2"
	//     Float: "1.5", "2.4"
	//     String: "abc", "defg"
	//     Boolean: "true", "false"
	//     IETF JSON from file: "@ap_config.json"
	Val string `json:"val"`
}

// TestResult contains the result of one gNMI test.
type TestResult struct {
	// Name is the name of the test.
	Name string
	// PassedNum is the number of passed test cases.
	PassedNum int
	// FailedNum is the number of failed test cases.
	FailedNum int
	// Details contains detailed test results of each test cases.
	Details []*TestCaseResult
}

// TestCaseResult contains the result of one gNMI test case.
type TestCaseResult struct {
	// Name is the test case name.
	Name string
	// Err is the error detected in this test case. nil if test passed.
	Err error
}
