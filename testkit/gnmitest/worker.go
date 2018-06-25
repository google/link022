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

// Package gnmitest contains functions related to gNMI tests.
// They test the functionalities of configuring or fetching telemetry data from AP devices through gNMI.
package gnmitest

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/google/gnxi/utils/xpath"
	"github.com/google/link022/testkit/common"
	"github.com/google/link022/testkit/util/gnmiutil"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// RunTest runs one gNMI test case.
// Args:
//   client: A gNMI client. It is used to send gNMI requests.
//   testCase: The target test case to run.
//   timeout: The timeout for each gNMI request. The test case failes if hitting timeout.
// Returns:
//   nil if test case passed. Otherwise, return the error with failure details.
func RunTest(client pb.GNMIClient, testCase *common.TestCase, timeout time.Duration) error {
	if client == nil {
		return errors.New("gNMI client is not available")
	}
	if testCase == nil {
		return errors.New("empty test case")
	}
	if len(testCase.OPs) == 0 {
		// Succeed if no operation specified in this test case.
		return nil
	}
	// Determine the test case type.
	switch testCase.OPs[0].Type {
	case common.OPReplace, common.OPUpdate, common.OPDelete:
		// This is a config test.
		return runConfigTest(client, testCase, timeout)
	case common.OPGet:
		// This is a state fetching test.
		return runStateTest(client, testCase, timeout)
	case common.OPSubscribe:
		return errors.New("not support telemetry streaming test cases")
	default:
		return fmt.Errorf("invalid operation type %s", testCase.OPs[0].Type)
	}
}

func runConfigTest(client pb.GNMIClient, testCase *common.TestCase, timeout time.Duration) error {
	// Generate the gNMI SetRequest containing all desired operations.
	setRequest, expectedVals, err := buildGNMISetRequest(testCase.OPs)
	if err != nil {
		return fmt.Errorf("unable to build gNMI SetRequest: %v.", err)
	}

	// Send the SetRequest.
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	setResponse, err := client.Set(ctx, setRequest)
	if err != nil {
		return fmt.Errorf("gNMI set request failed: %v", err)
	}

	// Check the response.
	if err := verifySetResponse(setResponse, expectedVals); err != nil {
		return err
	}

	// Check the pushed configuration updates are on device.
	return verifyConfiguration(client, expectedVals, timeout)
}

func runStateTest(client pb.GNMIClient, testCase *common.TestCase, timeout time.Duration) error {
	var desiredPaths []*pb.Path
	expectedVals := make(map[*pb.Path]*pb.TypedValue)
	for _, op := range testCase.OPs {
		if op.Type != common.OPGet {
			return fmt.Errorf("invalid operation type %s in gNMI tests, only allow %s", op.Type, common.OPGet)
		}
		pbPath, err := xpath.ToGNMIPath(op.Path)
		if err != nil {
			return err
		}
		pbVal, err := gnmiutil.ToPbVal(op.Val)
		if err != nil {
			return err
		}
		desiredPaths = append(desiredPaths, pbPath)
		expectedVals[pbPath] = pbVal
	}

	// Generate the gNMI GetRequest containing all desired paths.
	getRequest := &pb.GetRequest{
		Path:     desiredPaths,
		Encoding: pb.Encoding_JSON_IETF,
		UseModels: []*pb.ModelData{{
			Name:         "office-ap",
			Organization: "Google, Inc.",
			Version:      "0.1.0",
		}},
	}

	// Send gNMI GetRequest.
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	getResponse, err := client.Get(ctx, getRequest)
	if err != nil {
		return err
	}

	// Check the response.
	return verifyGetResponse(getResponse, expectedVals)
}

func buildGNMISetRequest(ops []*common.Operation) (*pb.SetRequest, map[*pb.Path]*pb.TypedValue, error) {
	var deleteList []*pb.Path
	var replaceList, updateList []*pb.Update
	expectedVals := make(map[*pb.Path]*pb.TypedValue)

	for _, op := range ops {
		pbPath, err := xpath.ToGNMIPath(op.Path)
		if err != nil {
			return nil, nil, err
		}
		switch op.Type {
		case common.OPReplace:
			pbVal, err := gnmiutil.ToPbVal(op.Val)
			if err != nil {
				return nil, nil, err
			}
			replaceList = append(replaceList, &pb.Update{Path: pbPath, Val: pbVal})
			expectedVals[pbPath] = pbVal
		case common.OPUpdate:
			pbVal, err := gnmiutil.ToPbVal(op.Val)
			if err != nil {
				return nil, nil, err
			}
			updateList = append(updateList, &pb.Update{Path: pbPath, Val: pbVal})
			expectedVals[pbPath] = pbVal
		case common.OPDelete:
			deleteList = append(deleteList, pbPath)
			expectedVals[pbPath] = nil
		default:
			return nil, nil, fmt.Errorf("invalid operation type %s for SET operation", op.Type)
		}
	}

	return &pb.SetRequest{
		Delete:  deleteList,
		Replace: replaceList,
		Update:  updateList,
	}, expectedVals, nil
}

func verifySetResponse(setResponse *pb.SetResponse, expectedVals map[*pb.Path]*pb.TypedValue) error {
	if len(setResponse.Response) != len(expectedVals) {
		return fmt.Errorf("incorrect response number in SetResponse, actual = %d, expected = %d", len(setResponse.Response), len(expectedVals))
	}

	prefix := setResponse.Prefix
	for _, updateResp := range setResponse.Response {
		targetFullPath := gnmiutil.GNMIFullPath(prefix, updateResp.Path)
		expectedVal, ok := fetchVal(expectedVals, targetFullPath)
		if !ok {
			return fmt.Errorf("unexpected path %v in SetResponse, waiting for %v", targetFullPath, expectedVals)
		}
		switch updateResp.Op {
		case pb.UpdateResult_DELETE:
			if expectedVal != nil {
				return fmt.Errorf("incorrect operation type %v on path %v in SetResponse, expected = %v", updateResp.Op, targetFullPath, pb.UpdateResult_DELETE)
			}
		case pb.UpdateResult_REPLACE, pb.UpdateResult_UPDATE:
			if expectedVal == nil {
				return fmt.Errorf("incorrect operation type %v on path %v in SetResponse, expected = %v/%v", updateResp.Op, targetFullPath, pb.UpdateResult_REPLACE, pb.UpdateResult_UPDATE)
			}
		default:
			return fmt.Errorf("invalid operation type %v in SetResponse", updateResp.Op)
		}
	}
	return nil
}

func fetchVal(expectedVals map[*pb.Path]*pb.TypedValue, targetPath *pb.Path) (*pb.TypedValue, bool) {
	for gnmiPath, expectedVal := range expectedVals {
		if gnmiutil.GNMIPathEquals(targetPath, gnmiPath) {
			return expectedVal, true
		}
	}
	return nil, false
}

func verifyConfiguration(client pb.GNMIClient, expectedVals map[*pb.Path]*pb.TypedValue, timeout time.Duration) error {
	for gnmiPath, expectedVal := range expectedVals {
		// Fetch the updated config from device.
		getRequest := &pb.GetRequest{
			Path:     []*pb.Path{gnmiPath},
			Encoding: pb.Encoding_JSON_IETF,
			UseModels: []*pb.ModelData{{
				Name:         "office-ap",
				Organization: "Google, Inc.",
				Version:      "0.1.0",
			}},
		}
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		getResponse, err := client.Get(ctx, getRequest)

		// Check deleted config. Verify they are actually removed.
		if expectedVal == nil && err == nil {
			return fmt.Errorf("still able to fetch the deleted config on %v", gnmiPath)
		}

		if err != nil {
			return fmt.Errorf("gNMI get request failed for path: %v", err)
		}

		// Check updated config is correct.
		if err := verifyGetResponse(getResponse, map[*pb.Path]*pb.TypedValue{gnmiPath: expectedVal}); err != nil {
			return err
		}
	}

	return nil
}

func verifyGetResponse(getResponse *pb.GetResponse, expectedVals map[*pb.Path]*pb.TypedValue) error {
	if len(getResponse.Notification) != len(expectedVals) {
		return fmt.Errorf("incorrect notification number in GetResponse, actual = %d, expected = %d", len(getResponse.Notification), len(expectedVals))
	}

	for _, notification := range getResponse.Notification {
		if len(notification.Delete) != 0 {
			return fmt.Errorf("incorrect Delete number in GetResponse, actual = %d, expected = 0", len(notification.Delete))
		}
		if len(notification.Update) != 1 {
			return fmt.Errorf("incorrect Update number in GetResponse, actual = %d, expected = 1", len(notification.Update))
		}

		update := notification.Update[0]
		updatedPath := gnmiutil.GNMIFullPath(notification.Prefix, update.Path)
		expectedVal, ok := fetchVal(expectedVals, updatedPath)
		if !ok {
			return fmt.Errorf("unexpected path %v in GetResponse, waiting for %v", updatedPath, expectedVals)
		}
		if err := gnmiutil.ValEqual(updatedPath, update.Val, expectedVal); err != nil {
			return err
		}
	}

	return nil
}
