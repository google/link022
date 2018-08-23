package gnmiutil

import (
	"testing"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	sampleConfigJSON = `{
   "radio":[
      {
         "config":{
            "channel":157,
            "channel-width":40,
            "enabled":true,
            "id":0,
            "operating-frequency":"openconfig-wifi-types:FREQ_5GHZ",
            "transmit-power":9
         },
         "id":0
      },
      {
         "config":{
            "operating-frequency":"openconfig-wifi-types:FREQ_2GHZ",
            "channel-width":20,
            "enabled":true,
            "dtp":false,
            "transmit-power":3,
            "id":1,
            "channel":6
         },
         "id":1
      }
   ]
}`
	sampleConfigStateJSON = `{
   "radio":[
      {
         "config":{
            "channel":157,
            "channel-width":40,
            "enabled":true,
            "id":0,
            "operating-frequency":"openconfig-wifi-types:FREQ_5GHZ",
            "transmit-power":9
         },
         "id":0,
         "state":{
            "base-radio-mac":"5c:5b:35:00:2b:00",
            "channel":157,
            "channel-width":40,
            "counters":{
               "noise-floor":-86
            },
            "enabled":true,
            "id":0,
            "operating-frequency":"openconfig-wifi-types:FREQ_5GHZ",
            "rx-noise-channel-utilization":19,
            "total-channel-utilization":20,
            "transmit-power":9
         }
      },
      {
         "config":{
            "channel":6,
            "channel-width":20,
            "dtp":false,
            "enabled":true,
            "id":1,
            "operating-frequency":"openconfig-wifi-types:FREQ_2GHZ",
            "transmit-power":3
         },
         "id":1,
         "state":{
            "base-radio-mac":"5c:5b:35:00:2a:f0",
            "channel":6,
            "channel-width":20,
            "counters":{
               "noise-floor":-73
            },
            "dtp":false,
            "enabled":true,
            "id":1,
            "operating-frequency":"openconfig-wifi-types:FREQ_2GHZ",
            "rx-noise-channel-utilization":82,
            "total-channel-utilization":83,
            "transmit-power":3
         }
      }
   ]
}`
)

func TestValEqual(t *testing.T) {
	testCases := []struct {
		actual   *pb.TypedValue
		expected *pb.TypedValue
		isEqual  bool
	}{
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_IntVal{
					IntVal: 6,
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_IntVal{
					IntVal: 6,
				},
			},
			isEqual: true,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_IntVal{
					IntVal: 6,
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_IntVal{
					IntVal: 10,
				},
			},
			isEqual: false,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_FloatVal{
					FloatVal: 3.33,
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_FloatVal{
					FloatVal: 3.33,
				},
			},
			isEqual: true,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_FloatVal{
					FloatVal: 3.33,
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_FloatVal{
					FloatVal: 6.666,
				},
			},
			isEqual: false,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_BoolVal{
					BoolVal: true,
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_BoolVal{
					BoolVal: true,
				},
			},
			isEqual: true,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_BoolVal{
					BoolVal: true,
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_BoolVal{
					BoolVal: false,
				},
			},
			isEqual: false,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_StringVal{
					StringVal: "One",
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_StringVal{
					StringVal: "One",
				},
			},
			isEqual: true,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_StringVal{
					StringVal: "One",
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_StringVal{
					StringVal: "Two",
				},
			},
			isEqual: false,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_JsonIetfVal{
					JsonIetfVal: []byte(sampleConfigStateJSON),
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_JsonIetfVal{
					JsonIetfVal: []byte(sampleConfigJSON),
				},
			},
			isEqual: true,
		},
		{
			actual: &pb.TypedValue{
				Value: &pb.TypedValue_BoolVal{
					BoolVal: true,
				},
			},
			expected: &pb.TypedValue{
				Value: &pb.TypedValue_StringVal{
					StringVal: "Two",
				},
			},
			isEqual: false,
		},
	}

	for _, testCase := range testCases {
		if err := ValEqual(nil, testCase.actual, testCase.expected); (err == nil) != testCase.isEqual {
			t.Errorf("incorrect matching result, actual = %v, expected = %v. (matching candidate: %v, %v)", err == nil, testCase.isEqual, testCase.actual, testCase.expected)
		}
	}
}
