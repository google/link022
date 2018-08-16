package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/golang/glog"
	"github.com/google/gnxi/utils/xpath"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/value"
)

const (
	testNanoTimeStamp    = 1529343009489648
	testGNMIPrefix       = "/p1/p2[name=abc]/p3"
	testGNMIPathInt      = "/a/b-b/c[id=0]/d/intval"
	testGNMIPathFloat    = "/a/b-b/c[id=0]/d/floatval"
	testGNMIPathString   = "/a/b-b/c[id=0]/d/stringval"
	testGNMIPathArray    = "a/c/d[id=1]/e/array"
	testGNMIPathIntArray = "a/c/d[id=1]/e/intarray"
	testGNMIPathJSON     = "a/c/d[id=1]/e/json"
	testGNMIPathJSONTree = "a/c/d[id=1]/e/jsontree"
	testGNMIPathJSONIETF = "/access-points/access-point[hostname=link022-pi-ap]/system/aaa/server-groups"
	testMetricIntVal     = 100
	testMetricDoubleVal  = float32(1.1)
	testMetricStringVal  = "test_val"
)

func newGNMINotification() []*gpb.Notification {
	intVal, err := value.FromScalar(int64(testMetricIntVal))
	if err != nil {
		log.Errorf("convert %v to GNMI value type failed", testMetricIntVal)
		return nil
	}
	floatVal, err := value.FromScalar(testMetricDoubleVal)
	if err != nil {
		log.Errorf("convert %v to GNMI value type failed", testMetricDoubleVal)
		return nil
	}
	stringVal, err := value.FromScalar(testMetricStringVal)
	if err != nil {
		log.Errorf("convert %v to GNMI value type failed", testMetricStringVal)
		return nil
	}
	arrayIntf := make([]interface{}, 3)
	arrayIntf[0] = testMetricIntVal
	arrayIntf[1] = testMetricDoubleVal
	arrayIntf[2] = testMetricStringVal
	arrayVal, err := value.FromScalar(arrayIntf)
	if err != nil {
		log.Errorf("convert %v to GNMI value type failed", arrayIntf)
		return nil
	}

	intArrayIntf := make([]interface{}, 3)
	intArrayIntf[0] = testMetricIntVal
	intArrayIntf[1] = testMetricIntVal
	intArrayIntf[2] = testMetricIntVal
	intArrayVal, err := value.FromScalar(intArrayIntf)
	if err != nil {
		log.Errorf("convert %v to GNMI value type failed", intArrayIntf)
		return nil
	}

	jsonBolb, err := json.Marshal(testMetricDoubleVal)
	if err != nil {
		log.Errorf("convert %v to JSON bolb type failed: %v", testMetricDoubleVal, err)
	}
	jsonVal := &gpb.TypedValue{Value: &gpb.TypedValue_JsonVal{JsonVal: jsonBolb}}
	jsonTree := make(map[string]interface{})
	jsonTree["number1"] = testMetricDoubleVal
	jsonTree["string"] = testMetricStringVal
	jsonTree["number2"] = testMetricIntVal
	jsonTreeBolb, err := json.Marshal(jsonTree)
	if err != nil {
		log.Errorf("convert %v to JSON bolb type failed: %v", jsonTree, err)
	}
	jsonTreeVal := &gpb.TypedValue{Value: &gpb.TypedValue_JsonVal{JsonVal: jsonTreeBolb}}

	jsonIETFString := `{  
		"server-group":[  
		   {  
			  "servers":{  
				 "server":[  
					{  
					   "address":"192.168.11.1",
					   "config":{  
						  "address":"192.168.11.1",
						  "timeout":5,
						  "name":"radius-server"
					   },
					   "radius":{  
						  "config":{  
							 "auth-port":1812,
							 "secret-key":"radiuspwd"
						  }
					   }
					}
				 ]
			  },
			  "name":"freeradius",
			  "config":{  
				 "name":"freeradius",
				 "type":"openconfig-aaa:RADIUS"
			  }
		   }
		]
	 }`
	jsonIETFVal := &gpb.TypedValue{Value: &gpb.TypedValue_JsonIetfVal{JsonIetfVal: []byte(jsonIETFString)}}

	prefixPath, err := xpath.ToGNMIPath(testGNMIPrefix)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPrefix)
		return nil
	}
	gnmiPathInt, err := xpath.ToGNMIPath(testGNMIPathInt)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPathInt)
		return nil
	}
	gnmiPathFloat, err := xpath.ToGNMIPath(testGNMIPathFloat)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPathFloat)
		return nil
	}
	gnmiPathString, err := xpath.ToGNMIPath(testGNMIPathString)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPathString)
		return nil
	}
	gnmiPathArray, err := xpath.ToGNMIPath(testGNMIPathArray)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPathArray)
		return nil
	}
	gnmiPathIntArray, err := xpath.ToGNMIPath(testGNMIPathIntArray)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPathIntArray)
		return nil
	}
	gnmiPathJSON, err := xpath.ToGNMIPath(testGNMIPathJSON)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPathJSON)
		return nil
	}
	gnmiPathJSONTree, err := xpath.ToGNMIPath(testGNMIPathJSONTree)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPathJSONTree)
		return nil
	}
	gnmiPathJSONIETF, err := xpath.ToGNMIPath(testGNMIPathJSONIETF)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed", testGNMIPathJSONIETF)
		return nil
	}

	noti :=
		&gpb.Notification{
			Timestamp: testNanoTimeStamp,
			Prefix:    prefixPath,
			Update: []*gpb.Update{
				{
					Path: gnmiPathInt,
					Val:  intVal,
				},
				{
					Path: gnmiPathFloat,
					Val:  floatVal,
				},
				{
					Path: gnmiPathString,
					Val:  stringVal,
				},
				{
					Path: gnmiPathArray,
					Val:  arrayVal,
				},
				{
					Path: gnmiPathIntArray,
					Val:  intArrayVal,
				},
				{
					Path: gnmiPathJSON,
					Val:  jsonVal,
				},
				{
					Path: gnmiPathJSONTree,
					Val:  jsonTreeVal,
				},
			},
		}
	notiJSONIETF :=
		&gpb.Notification{
			Timestamp: testNanoTimeStamp,
			Prefix:    nil,
			Update: []*gpb.Update{
				{
					Path: gnmiPathJSONIETF,
					Val:  jsonIETFVal,
				},
			},
		}
	return []*gpb.Notification{noti, notiJSONIETF}
}

// TestPrometheusHandler will test HTTP handler of this Prometheus exporter.
// Because exporter will also expose some runtime information, this test only
// check input test metrics.
func TestPrometheusHandler(t *testing.T) {
	notis := newGNMINotification()
	targetState := &TargetState{
		state: notis[0],
	}
	gnmiExport := newAPStateCollector(targetState)
	prometheus.MustRegister(gnmiExport)

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("Initialize test request failed: %v", err)
	}
	responRecord := httptest.NewRecorder()

	handler := promhttp.Handler()
	handler.ServeHTTP(responRecord, req)
	if status := responRecord.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expectedIntVal := "# HELP p1:p2:p3:a:b_b:c:d:intval int64 type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:b_b:c:d:intval gauge\n" +
		"p1:p2:p3:a:b_b:c:d:intval{c_id=\"0\",p2_name=\"abc\"} 100"
	if strings.Index(responRecord.Body.String(), expectedIntVal) == -1 {
		t.Errorf("Can't find int test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedIntVal)
	}
	expectedFloatVal := "# HELP p1:p2:p3:a:b_b:c:d:floatval float32 type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:b_b:c:d:floatval gauge\n" +
		"p1:p2:p3:a:b_b:c:d:floatval{c_id=\"0\",p2_name=\"abc\"} 1.1"
	if strings.Index(responRecord.Body.String(), expectedFloatVal) == -1 {
		t.Errorf("Can't find float test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedFloatVal)
	}
	expectedStringVal := "# HELP p1:p2:p3:a:b_b:c:d:stringval string type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:b_b:c:d:stringval untyped\n" +
		"p1:p2:p3:a:b_b:c:d:stringval{c_id=\"0\",metric_value=\"test_val\",p2_name=\"abc\"} 0"
	if strings.Index(responRecord.Body.String(), expectedStringVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedStringVal)
	}
	expectedArrayVal := "# HELP p1:p2:p3:a:c:d:e:array string array type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:c:d:e:array untyped\n" +
		"p1:p2:p3:a:c:d:e:array{array_index=\"0\",d_id=\"1\",metric_value=\"100\",p2_name=\"abc\"} 0\n" +
		"p1:p2:p3:a:c:d:e:array{array_index=\"1\",d_id=\"1\",metric_value=\"1.1\",p2_name=\"abc\"} 0\n" +
		"p1:p2:p3:a:c:d:e:array{array_index=\"2\",d_id=\"1\",metric_value=\"test_val\",p2_name=\"abc\"} 0\n"
	if strings.Index(responRecord.Body.String(), expectedArrayVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedArrayVal)
	}
	expectedIntArrayVal := "# HELP p1:p2:p3:a:c:d:e:intarray int64 array type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:c:d:e:intarray gauge\n" +
		"p1:p2:p3:a:c:d:e:intarray{array_index=\"0\",d_id=\"1\",p2_name=\"abc\"} 100\n" +
		"p1:p2:p3:a:c:d:e:intarray{array_index=\"1\",d_id=\"1\",p2_name=\"abc\"} 100\n" +
		"p1:p2:p3:a:c:d:e:intarray{array_index=\"2\",d_id=\"1\",p2_name=\"abc\"} 100\n"
	if strings.Index(responRecord.Body.String(), expectedIntArrayVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedIntArrayVal)
	}
	expectedJSONVal := "# HELP p1:p2:p3:a:c:d:e:json float32 type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:c:d:e:json gauge\n" +
		"p1:p2:p3:a:c:d:e:json{d_id=\"1\",p2_name=\"abc\"} 1.1"
	if strings.Index(responRecord.Body.String(), expectedJSONVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONVal)
	}
	expectedJSONTreeNumber1Val := "# HELP p1:p2:p3:a:c:d:e:jsontree:number1 float32 type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:c:d:e:jsontree:number1 gauge\n" +
		"p1:p2:p3:a:c:d:e:jsontree:number1{d_id=\"1\",p2_name=\"abc\"} 1.1"
	if strings.Index(responRecord.Body.String(), expectedJSONTreeNumber1Val) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONTreeNumber1Val)
	}
	expectedJSONTreeNumber2Val := "# HELP p1:p2:p3:a:c:d:e:jsontree:number2 float32 type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:c:d:e:jsontree:number2 gauge\n" +
		"p1:p2:p3:a:c:d:e:jsontree:number2{d_id=\"1\",p2_name=\"abc\"} 100"
	if strings.Index(responRecord.Body.String(), expectedJSONTreeNumber2Val) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONTreeNumber2Val)
	}
	expectedJSONTreeStringVal := " HELP p1:p2:p3:a:c:d:e:jsontree:string string type gNMI metric\n" +
		"# TYPE p1:p2:p3:a:c:d:e:jsontree:string untyped\n" +
		"p1:p2:p3:a:c:d:e:jsontree:string{d_id=\"1\",metric_value=\"test_val\",p2_name=\"abc\"} 0"
	if strings.Index(responRecord.Body.String(), expectedJSONTreeStringVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONTreeStringVal)
	}
	prometheus.Unregister(gnmiExport)
}

// TestPrometheusHandlerJSONIETF will test HTTP handler of this Prometheus exporter.
// Input gNMI notification only contain JSON_IETF type data.
func TestPrometheusHandlerJSONIETF(t *testing.T) {
	notis := newGNMINotification()
	targetState := &TargetState{
		state: notis[1],
	}
	gnmiExport := newAPStateCollector(targetState)
	prometheus.MustRegister(gnmiExport)

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("Initialize test request failed: %v", err)
	}
	responRecord := httptest.NewRecorder()

	handler := promhttp.Handler()
	handler.ServeHTTP(responRecord, req)
	if status := responRecord.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expectedJSONIETFVal := "# HELP access_points:access_point:system:aaa:server_groups:server_group:config:name string type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:config:name untyped\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:config:name{access_point_hostname=\"link022-pi-ap\",metric_value=\"freeradius\",server_group_name=\"freeradius\"} 0"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	expectedJSONIETFVal = "# HELP access_points:access_point:system:aaa:server_groups:server_group:config:type string type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:config:type untyped\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:config:type{access_point_hostname=\"link022-pi-ap\",metric_value=\"RADIUS\",server_group_name=\"freeradius\"} 0"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	expectedJSONIETFVal = "# HELP access_points:access_point:system:aaa:server_groups:server_group:name string type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:name untyped\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:name{access_point_hostname=\"link022-pi-ap\",metric_value=\"freeradius\",server_group_name=\"freeradius\"} 0"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	expectedJSONIETFVal = "# HELP access_points:access_point:system:aaa:server_groups:server_group:servers:server:address string type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:servers:server:address untyped\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:servers:server:address{access_point_hostname=\"link022-pi-ap\",metric_value=\"192.168.11.1\",server_address=\"192.168.11.1\",server_group_name=\"freeradius\"} 0"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	expectedJSONIETFVal = "# HELP access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:address string type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:address untyped\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:address{access_point_hostname=\"link022-pi-ap\",metric_value=\"192.168.11.1\",server_address=\"192.168.11.1\",server_group_name=\"freeradius\"} 0"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	expectedJSONIETFVal = "# HELP access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:name string type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:name untyped\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:name{access_point_hostname=\"link022-pi-ap\",metric_value=\"radius-server\",server_address=\"192.168.11.1\",server_group_name=\"freeradius\"} 0"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	expectedJSONIETFVal = "# HELP access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:timeout uint64 type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:timeout gauge\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:servers:server:config:timeout{access_point_hostname=\"link022-pi-ap\",server_address=\"192.168.11.1\",server_group_name=\"freeradius\"} 5"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	expectedJSONIETFVal = "# HELP access_points:access_point:system:aaa:server_groups:server_group:servers:server:radius:config:auth_port uint64 type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:servers:server:radius:config:auth_port gauge\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:servers:server:radius:config:auth_port{access_point_hostname=\"link022-pi-ap\",server_address=\"192.168.11.1\",server_group_name=\"freeradius\"} 1812"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	expectedJSONIETFVal = "# HELP access_points:access_point:system:aaa:server_groups:server_group:servers:server:radius:config:secret_key string type gNMI metric\n" +
		"# TYPE access_points:access_point:system:aaa:server_groups:server_group:servers:server:radius:config:secret_key untyped\n" +
		"access_points:access_point:system:aaa:server_groups:server_group:servers:server:radius:config:secret_key{access_point_hostname=\"link022-pi-ap\",metric_value=\"radiuspwd\",server_address=\"192.168.11.1\",server_group_name=\"freeradius\"} 0"
	if strings.Index(responRecord.Body.String(), expectedJSONIETFVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedJSONIETFVal)
	}
	prometheus.Unregister(gnmiExport)
}
