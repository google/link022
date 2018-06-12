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

// Package gnmiconfig contains functions related to gNMI config.
// They test the functionalities of configuring AP devices through gNMI.
package gnmiconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/google/gnxi/utils/xpath"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// OPType is the type of gNMI SET operation.
type OPType string

const (
	// OPReplace is the gNMI replace operation.
	OPReplace OPType = "replace"
	// OPUpdate is the gNMI update operation.
	OPUpdate OPType = "update"
	// OPDelete is the gNMI delete operation.
	OPDelete OPType = "delete"
)

// TestCase describes a config-related gNMI test cases.
type TestCase struct {
	// Name is the test case name.
	Name string `json:"name"`
	// Description is the detail description of this test case.
	Description string `json:"description"`
	// OPs contains a list of operations need to be processed in this test case.
	// All operations are processed in one single gNMI SetRequest.
	OPs []*operation `json:"ops"`
}

type operation struct {
	// Type is the gNMI SET operation type.
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

// RunTest runs one gNMI-config test case.
// Args:
//   client: A gNMI client. It is used to send gNMI requests.
//   testCase: The target test case to run.
//   timeout: The timeout for each gNMI request. The test case failes if hitting timeout.
// Returns:
//   nil if test case passed. Otherwise, return the error with failure details.
func RunTest(client pb.GNMIClient, testCase *TestCase, timeout time.Duration) error {
	if client == nil {
		return errors.New("gNMI client is not available")
	}
	if testCase == nil {
		return errors.New("empty test case")
	}

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

func buildGNMISetRequest(ops []*operation) (*pb.SetRequest, map[*pb.Path]*pb.TypedValue, error) {
	var deleteList []*pb.Path
	var replaceList, updateList []*pb.Update
	expectedVals := make(map[*pb.Path]*pb.TypedValue)

	for _, op := range ops {
		switch op.Type {
		case OPReplace:
			pbUpdate, err := buildPbUpdate(op)
			if err != nil {
				return nil, nil, err
			}
			replaceList = append(replaceList, pbUpdate)
			expectedVals[pbUpdate.Path] = pbUpdate.Val
		case OPUpdate:
			pbUpdate, err := buildPbUpdate(op)
			if err != nil {
				return nil, nil, err
			}
			updateList = append(updateList, pbUpdate)
			expectedVals[pbUpdate.Path] = pbUpdate.Val
		case OPDelete:
			pbPath, err := xpath.ToGNMIPath(op.Path)
			if err != nil {
				return nil, nil, err
			}
			deleteList = append(deleteList, pbPath)
			expectedVals[pbPath] = nil
		default:
			return nil, nil, fmt.Errorf("unknown operation type %s", op.Type)
		}
	}

	return &pb.SetRequest{
		Delete:  deleteList,
		Replace: replaceList,
		Update:  updateList,
	}, expectedVals, nil
}

func buildPbUpdate(op *operation) (*pb.Update, error) {
	pbPath, err := xpath.ToGNMIPath(op.Path)
	if err != nil {
		return nil, err
	}

	var pbVal *pb.TypedValue
	if op.Val[0] == '@' {
		jsonFile := op.Val[1:]
		jsonConfig, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("cannot read data from file %v: %v", jsonFile, err)
		}
		jsonConfig = bytes.Trim(jsonConfig, " \r\n\t")
		pbVal = &pb.TypedValue{
			Value: &pb.TypedValue_JsonIetfVal{
				JsonIetfVal: jsonConfig,
			},
		}
	} else {
		if strVal, err := strconv.Unquote(op.Val); err == nil {
			pbVal = &pb.TypedValue{
				Value: &pb.TypedValue_StringVal{
					StringVal: strVal,
				},
			}
		} else {
			if intVal, err := strconv.ParseInt(op.Val, 10, 64); err == nil {
				pbVal = &pb.TypedValue{
					Value: &pb.TypedValue_IntVal{
						IntVal: intVal,
					},
				}
			} else if floatVal, err := strconv.ParseFloat(op.Val, 32); err == nil {
				pbVal = &pb.TypedValue{
					Value: &pb.TypedValue_FloatVal{
						FloatVal: float32(floatVal),
					},
				}
			} else if boolVal, err := strconv.ParseBool(op.Val); err == nil {
				pbVal = &pb.TypedValue{
					Value: &pb.TypedValue_BoolVal{
						BoolVal: boolVal,
					},
				}
			} else {
				pbVal = &pb.TypedValue{
					Value: &pb.TypedValue_StringVal{
						StringVal: op.Val,
					},
				}
			}
		}
	}
	return &pb.Update{Path: pbPath, Val: pbVal}, nil
}

// gnmiFullPath builds the full path from the prefix and path.
func gnmiFullPath(prefix, path *pb.Path) *pb.Path {
	if prefix == nil {
		return path
	}

	fullPath := &pb.Path{Origin: path.Origin}
	if path.GetElement() != nil {
		fullPath.Element = append(prefix.GetElement(), path.GetElement()...)
	}
	if path.GetElem() != nil {
		fullPath.Elem = append(prefix.GetElem(), path.GetElem()...)
	}
	return fullPath
}

func verifySetResponse(setResponse *pb.SetResponse, expectedVals map[*pb.Path]*pb.TypedValue) error {
	if len(setResponse.Response) != len(expectedVals) {
		return fmt.Errorf("incorrect response number in SetResponse, actual = %d, expected = %d", len(setResponse.Response), len(expectedVals))
	}

	prefix := setResponse.Prefix
	for _, updateResp := range setResponse.Response {
		targetFullPath := gnmiFullPath(prefix, updateResp.Path)
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
		if reflect.DeepEqual(targetPath, gnmiPath) {
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
		if err := verifyGetResponse(getResponse, gnmiPath, expectedVal); err != nil {
			return err
		}
	}

	return nil
}

func verifyGetResponse(getResponse *pb.GetResponse, gnmiPath *pb.Path, expectedVal *pb.TypedValue) error {
	if len(getResponse.Notification) != 1 {
		return fmt.Errorf("incorrect notification number in GetResponse, actual = %d, expected = 1", len(getResponse.Notification))
	}

	notification := getResponse.Notification[0]
	if len(notification.Delete) != 0 {
		return fmt.Errorf("incorrect Delete number in GetResponse, actual = %d, expected = 0", len(notification.Delete))
	}
	if len(notification.Update) != 1 {
		return fmt.Errorf("incorrect Update number in GetResponse, actual = %d, expected = 1", len(notification.Update))
	}

	update := notification.Update[0]
	updatedPath := gnmiFullPath(notification.Prefix, update.Path)
	if !reflect.DeepEqual(updatedPath, gnmiPath) {
		return fmt.Errorf("incorrect gnmi path in GetResponse, actual = %v, expected = %v", updatedPath, gnmiPath)
	}

	return valEqual(gnmiPath, update.Val, expectedVal)
}

func valEqual(gnmiPath *pb.Path, actual *pb.TypedValue, expected *pb.TypedValue) error {
	actualValue := actual.Value
	expectedValue := expected.Value

	// JSON value.
	actualJsonValue, okA := actualValue.(*pb.TypedValue_JsonIetfVal)
	expectedJsonValue, okE := expectedValue.(*pb.TypedValue_JsonIetfVal)
	if okA && okE {
		var actualJson, expectedJson interface{}
		if err := json.Unmarshal(actualJsonValue.JsonIetfVal, &actualJson); err != nil {
			return fmt.Errorf("invalid value %v: %v", string(actualJsonValue.JsonIetfVal), err)
		}
		if err := json.Unmarshal(expectedJsonValue.JsonIetfVal, &expectedJson); err != nil {
			return fmt.Errorf("invalid value %v: %v", string(expectedJsonValue.JsonIetfVal), err)
		}
		if !reflect.DeepEqual(actualJson, expectedJson) {
			return fmt.Errorf("incorrect json config value on %v, actual = %v, expected = %v", gnmiPath, string(actualJsonValue.JsonIetfVal), string(expectedJsonValue.JsonIetfVal))
		}
		return nil
	}

	// Other value types.

	// No need to check uint64/int64 type matching.
	if uintValue, ok := actualValue.(*pb.TypedValue_UintVal); ok {
		actualValue = &pb.TypedValue_IntVal{
			IntVal: int64(uintValue.UintVal),
		}
	}
	if !reflect.DeepEqual(actualValue, expectedValue) {
		return fmt.Errorf("incorrect config on %v, actual = %v, expected = %v", gnmiPath, actualValue, expectedValue)
	}
	return nil
}
