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

// Package gnmiutil contains helper functions related to gNMI.
package gnmiutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// GNMIFullPath builds the full path from the prefix and path.
func GNMIFullPath(prefix, path *pb.Path) *pb.Path {
	if prefix == nil {
		return path
	}

	fullPath := &pb.Path{Origin: path.Origin}
	if path.GetElem() != nil {
		fullPath.Elem = append(prefix.GetElem(), path.GetElem()...)
	}
	return fullPath
}

// ToPbVal convert string to TypedValue defined in gNMI proto.
// Supported types:
//     Integer: "1", "2"
//     Float: "1.5", "2.4"
//     String: "abc", "defg"
//     Boolean: "true", "false"
//     IETF JSON from file: "@ap_config.json"
func ToPbVal(stringVal string) (*pb.TypedValue, error) {
	if stringVal[0] == '@' {
		jsonFile := stringVal[1:]
		jsonConfig, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("cannot read data from file %v: %v", jsonFile, err)
		}
		jsonConfig = bytes.Trim(jsonConfig, " \r\n\t")
		return &pb.TypedValue{
			Value: &pb.TypedValue_JsonIetfVal{
				JsonIetfVal: jsonConfig,
			},
		}, nil
	}

	if strVal, err := strconv.Unquote(stringVal); err == nil {
		return &pb.TypedValue{
			Value: &pb.TypedValue_StringVal{
				StringVal: strVal,
			},
		}, nil
	}

	if intVal, err := strconv.ParseInt(stringVal, 10, 64); err == nil {
		return &pb.TypedValue{
			Value: &pb.TypedValue_IntVal{
				IntVal: intVal,
			},
		}, nil
	}

	if floatVal, err := strconv.ParseFloat(stringVal, 32); err == nil {
		return &pb.TypedValue{
			Value: &pb.TypedValue_FloatVal{
				FloatVal: float32(floatVal),
			},
		}, nil
	}

	if boolVal, err := strconv.ParseBool(stringVal); err == nil {
		return &pb.TypedValue{
			Value: &pb.TypedValue_BoolVal{
				BoolVal: boolVal,
			},
		}, nil
	}

	return &pb.TypedValue{
		Value: &pb.TypedValue_StringVal{
			StringVal: stringVal,
		},
	}, nil
}

// GNMIPathEquals checks whether the two given gNMI path equal to each other.
func GNMIPathEquals(actual, expected *pb.Path) bool {
	if len(actual.Elem) != len(expected.Elem) {
		return false
	}

	for i := 0; i < len(actual.Elem); i++ {
		actualElem := actual.Elem[i]
		expectedElem := expected.Elem[i]
		if actualElem.Name != expectedElem.Name || !reflect.DeepEqual(actualElem.Key, expectedElem.Key) {
			return false
		}
	}

	return true
}

// ValEqual checks whether the given two input value equal.
// It returns error if values are not equal.
func ValEqual(gnmiPath *pb.Path, actual *pb.TypedValue, expected *pb.TypedValue) error {
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
