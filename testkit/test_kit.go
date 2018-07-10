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

// The test_kit program is a tool that tests gNMI functionalities of an AP device.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/golang/glog"
	"google.golang.org/grpc"

	"github.com/google/gnxi/utils/credentials"

	"github.com/google/link022/testkit/common"
	"github.com/google/link022/testkit/gnmitest"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	gnmiTests  arrayFlags
	targetAddr = flag.String("target_addr", "localhost:10161", "The target address in the format of host:port")
	targetName = flag.String("target_name", "hostname.com", "The target name used to verify the hostname returned by TLS handshake")
	timeout    = flag.Duration("time_out", 30*time.Second, "Timeout for each request, 30 seconds by default")
	pauseMode  = flag.Bool("pause_mode", false, "Pause after each test case")
)

func loadTests(testFiles []string) ([]*common.GNMITest, error) {
	var tests []*common.GNMITest
	for _, testFile := range testFiles {
		testContent, err := ioutil.ReadFile(testFile)
		if err != nil {
			return nil, err
		}
		test := &common.GNMITest{}
		if err := json.Unmarshal(testContent, test); err != nil {
			return nil, err
		}
		tests = append(tests, test)
		log.Infof("Loaded [%s].", test.Name)
	}
	return tests, nil
}

func runTest(client pb.GNMIClient, gNMITest *common.GNMITest, timeout time.Duration) *common.TestResult {
	var testCaseResults []*common.TestCaseResult

	// Run gNMI config tests.
	log.Infof("Running [%s].", gNMITest.Name)
	var passedNum, failedNum int
	totalNum := len(gNMITest.GNMITestCase)
	for i, testcase := range gNMITest.GNMITestCase {
		log.Infof("Started [%s].", testcase.Name)
		err := gnmitest.RunTest(client, testcase, timeout)
		if err != nil {
			failedNum += 1
			log.Errorf("[%d/%d] [%s] failed: %v.", i+1, totalNum, testcase.Name, err)
		} else {
			passedNum += 1
			log.Infof("[%d/%d] [%s] succeeded.", i+1, totalNum, testcase.Name)
		}

		result := &common.TestCaseResult{
			Name: testcase.Name,
			Err:  err,
		}
		testCaseResults = append(testCaseResults, result)

		if *pauseMode && i < totalNum {
			reader := bufio.NewReader(os.Stdin)
			// Pause until user triggers next test case manually.
			fmt.Println("Press ENTER to start the next test case.")
			_, _ = reader.ReadString('\n')
		}
	}

	if failedNum > 0 {
		log.Errorf("[%s] failed.", gNMITest.Name)
	} else {
		log.Infof("[%s] succeeded.", gNMITest.Name)
	}

	return &common.TestResult{
		Name:      gNMITest.Name,
		PassedNum: passedNum,
		FailedNum: failedNum,
		Details:   testCaseResults,
	}
}

func resultString(passed bool) string {
	if passed {
		return "PASS"
	}
	return "FAIL"
}

func printResult(results []*common.TestResult) {
	fmt.Println("=Test results=")
	// Print details of each test.
	for _, test := range results {
		fmt.Println("--------------------")
		fmt.Printf("[%s] %s\n", resultString(test.FailedNum == 0), test.Name)
		for _, testCase := range test.Details {
			passed := testCase.Err == nil
			fmt.Printf("|-[%s] %s\n", resultString(passed), testCase.Name)
			if !passed {
				fmt.Printf(" \\- %v\n", testCase.Err)
			}
		}
	}
	// Print test result summay.
	fmt.Println("--------------------")
	for _, test := range results {
		fmt.Printf("[%s] [%s] Passed - %d, Failed - %d\n", resultString(test.FailedNum == 0), test.Name, test.PassedNum, test.FailedNum)
	}
}

func main() {
	flag.Var(&gnmiTests, "test_file", "The file containing gNMI test.")
	flag.Parse()

	log.Info("Test kit started.")

	// Load test cases.
	tests, err := loadTests(gnmiTests)
	if err != nil {
		log.Fatalf("Failed to load tests. Error: %v.", err)
	}
	log.Infof("Loaded %d test files..", len(tests))

	// Create gNMI client.
	opts := credentials.ClientCredentials(*targetName)
	conn, err := grpc.Dial(*targetAddr, opts...)
	if err != nil {
		log.Fatalf("Dialing to %q failed: %v", *targetAddr, err)
	}
	defer conn.Close()
	client := pb.NewGNMIClient(conn)

	// Run all tests.
	var results []*common.TestResult
	for _, test := range tests {
		results = append(results, runTest(client, test, *timeout))
	}

	// Print out the result.
	printResult(results)
}
