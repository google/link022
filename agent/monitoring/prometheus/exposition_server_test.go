package main

import (
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
	testMetricIntVal     = 100
	testMetricDoubleVal  = float32(1.1)
	testMetricStringVal  = "test_val"
)

func newGNMINotification() *gpb.Notification {
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
	noti :=
		&gpb.Notification{
			Timestamp: testNanoTimeStamp,
			Prefix:    prefixPath,
			Update: []*gpb.Update{
				&gpb.Update{
					Path: gnmiPathInt,
					Val:  intVal,
				},
				&gpb.Update{
					Path: gnmiPathFloat,
					Val:  floatVal,
				},
				&gpb.Update{
					Path: gnmiPathString,
					Val:  stringVal,
				},
				&gpb.Update{
					Path: gnmiPathArray,
					Val:  arrayVal,
				},
				&gpb.Update{
					Path: gnmiPathIntArray,
					Val:  intArrayVal,
				},
			},
		}
	return noti
}

// TestPrometheusHandler will test HTTP handler of this Prometheus exporter.
// Because exporter will also expose some runtime infomation, this test only
// check input test metrics.
func TestPrometheusHandler(t *testing.T) {
	targetState := &TargetState{
		state: newGNMINotification(),
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
	expectedIntVal := "# HELP p1:p2:p3a:b_b:c:d:intval int64 type gNMI metric\n" +
		"# TYPE p1:p2:p3a:b_b:c:d:intval gauge\n" +
		"p1:p2:p3a:b_b:c:d:intval{c_id=\"0\",p2_name=\"abc\"} 100"
	if strings.Index(responRecord.Body.String(), expectedIntVal) == -1 {
		t.Errorf("Can't find int test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedIntVal)
	}
	expectedFloatVal := "# HELP p1:p2:p3a:b_b:c:d:floatval float32 type gNMI metric\n" +
		"# TYPE p1:p2:p3a:b_b:c:d:floatval gauge\n" +
		"p1:p2:p3a:b_b:c:d:floatval{c_id=\"0\",p2_name=\"abc\"} 1.1"
	if strings.Index(responRecord.Body.String(), expectedFloatVal) == -1 {
		t.Errorf("Can't find float test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedFloatVal)
	}
	expectedStringVal := "# HELP p1:p2:p3a:b_b:c:d:stringval string type gNMI metric\n" +
		"# TYPE p1:p2:p3a:b_b:c:d:stringval untyped\n" +
		"p1:p2:p3a:b_b:c:d:stringval{c_id=\"0\",metric_value=\"test_val\",p2_name=\"abc\"} 0"
	if strings.Index(responRecord.Body.String(), expectedStringVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedStringVal)
	}
	expectedArrayVal := "# HELP p1:p2:p3a:c:d:e:array string array type gNMI metric\n" +
		"# TYPE p1:p2:p3a:c:d:e:array untyped\n" +
		"p1:p2:p3a:c:d:e:array{array_index=\"0\",d_id=\"1\",metric_value=\"100\",p2_name=\"abc\"} 0\n" +
		"p1:p2:p3a:c:d:e:array{array_index=\"1\",d_id=\"1\",metric_value=\"1.1\",p2_name=\"abc\"} 0\n" +
		"p1:p2:p3a:c:d:e:array{array_index=\"2\",d_id=\"1\",metric_value=\"test_val\",p2_name=\"abc\"} 0\n"
	if strings.Index(responRecord.Body.String(), expectedArrayVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedArrayVal)
	}
	expectedIntArrayVal := "# HELP p1:p2:p3a:c:d:e:intarray int64 array type gNMI metric\n" +
		"# TYPE p1:p2:p3a:c:d:e:intarray gauge\n" +
		"p1:p2:p3a:c:d:e:intarray{array_index=\"0\",d_id=\"1\",p2_name=\"abc\"} 100\n" +
		"p1:p2:p3a:c:d:e:intarray{array_index=\"1\",d_id=\"1\",p2_name=\"abc\"} 100\n" +
		"p1:p2:p3a:c:d:e:intarray{array_index=\"2\",d_id=\"1\",p2_name=\"abc\"} 100\n"
	if strings.Index(responRecord.Body.String(), expectedIntArrayVal) == -1 {
		t.Errorf("Can't find string test metric in exposed metrics: got %v want %v",
			responRecord.Body.String(), expectedIntArrayVal)
	}
}
