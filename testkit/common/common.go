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

import (
	"github.com/google/link022/testkit/gnmiconfig"
)

// GNMITest is top-level model of gNMI test.
type GNMITest struct {
	// Name is the test name.
	Name string `json:"name"`
	// Description is the detail description of this test.
	Description string `json:"description"`
	// ConfigTests is the list of config-related test cases to run in this test.
	ConfigTests []*gnmiconfig.TestCase `json:"config_tests"`

	// TODO: Add state-related test cases.
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
