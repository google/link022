package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/openconfig/ygot/experimental/ygotutils"

	"github.com/openconfig/gnmi/value"
	"github.com/openconfig/ygot/ygot"

	log "github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/google/gnxi/utils/credentials"
	"github.com/google/gnxi/utils/xpath"
	"github.com/google/link022/generated/ocstruct"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	cpb "google.golang.org/genproto/googleapis/rpc/code"
)

const (
	apStatsExportingDelay = 15 * time.Second
	timeOut               = 10 * time.Second
	statusPath            = "/"
	// ArrayIndexLabel is the label used to indicate index in original array.
	// This is the key of the label.
	ArrayIndexLabel = "array_index"
)

// TargetState contain current state of the assigned ap device
// State is saved as gNMI Notification. It is more easier to extract metrics.
type TargetState struct {
	state *gpb.Notification
	mutex sync.RWMutex
}

func (s *TargetState) updateState(state *gpb.Notification) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state = state
}

func (s *TargetState) currentState() *gpb.Notification {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state
}

//monitoringAPStats
func monitoringAPStats(ctx context.Context, targetAddress string, targetName string, targetState *TargetState) {
	opts := credentials.ClientCredentials(targetName)
	conn, err := grpc.Dial(targetAddress, opts...)
	if err != nil {
		log.Errorf("Dialing to %q failed: %v", targetAddress, err)
		return
	}
	defer conn.Close()

	cli := gpb.NewGNMIClient(conn)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(apStatsExportingDelay):
		}
		var pathList []*gpb.Path
		gpbPath, err := xpath.ToGNMIPath(statusPath)
		pathList = append(pathList, gpbPath)

		getRequest := &gpb.GetRequest{
			Path: pathList,
			//Encoding: gpb.Encoding(encoding),
		}

		getResponse, err := cli.Get(ctx, getRequest)
		if err != nil {
			log.Errorf("Get failed: %v", err)
		}
		var newestID int
		var newestTimeStamp int64
		newestTimeStamp = 0
		for idx, notif := range getResponse.GetNotification() {
			timeStamp := notif.GetTimestamp()
			if newestTimeStamp < timeStamp {
				newestID = idx
				newestTimeStamp = timeStamp
			}
		}
		if newestTimeStamp > 0 {
			targetState.updateState(getResponse.GetNotification()[newestID])
		}
	}
}

// APStateCollector is the collector to get current AP status
type APStateCollector struct {
	state *TargetState
}

func newAPStateCollector(s *TargetState) *APStateCollector {
	if s == nil {
		return nil
	}
	return &APStateCollector{state: s}
}

// gNMIPathtoString splits GNMI Path into a path string and a label set.
// Prometheus metric name can only contain letters, number, underscore and colon.
// Prometheus label name can only contain letters, number and underscore
func gNMIPathtoString(in *gpb.Path) (string, map[string]string) {
	if in == nil {
		return "", nil
	}
	path := ""
	labels := make(map[string]string)
	for idx, ele := range in.Elem {
		elementName := strings.Replace(ele.Name, "-", "_", -1)
		elementName = strings.Replace(elementName, "/", ":", -1)
		if idx == 0 {
			path += elementName
		} else {
			path += ":" + elementName
		}
		for k, v := range ele.Key {
			k = strings.Replace(k, "-", "_", -1)
			k = strings.Replace(k, "/", "_", -1)
			labels[elementName+"_"+k] = v
		}
	}
	return path, labels
}

// check if all elements in input array have same type
func checkSingleType(leafList []*gpb.TypedValue) bool {
	if len(leafList) == 0 {
		return true
	}
	for i := 1; i < len(leafList); i++ {
		if reflect.TypeOf(leafList[i].Value) != reflect.TypeOf(leafList[0].Value) {
			return false
		}
	}
	return true
}

func typedValueToScalar(tv *gpb.TypedValue) (interface{}, error) {
	var i interface{}
	switch tv.Value.(type) {
	case *gpb.TypedValue_DecimalVal,
		*gpb.TypedValue_FloatVal,
		*gpb.TypedValue_StringVal,
		*gpb.TypedValue_IntVal,
		*gpb.TypedValue_UintVal,
		*gpb.TypedValue_BoolVal:
		val, err := value.ToScalar(tv)
		if err != nil {
			log.Errorf("convert gNMI %T type to scalar type failed: %v", tv.Value, err)
			return nil, err
		}
		i = val
	case *gpb.TypedValue_LeaflistVal:
		elems := tv.GetLeaflistVal().GetElement()
		ss := make([]interface{}, len(elems))
		if checkSingleType(elems) {
			for x, e := range elems {
				v, err := typedValueToScalar(e)
				if err != nil {
					return nil, fmt.Errorf("convert gNMI %T type to scalar type failed: %v", e.Value, err)
				}
				ss[x] = v
			}
		} else {
			for x, e := range elems {
				scalarEle, err := value.ToScalar(e)
				if err != nil {
					return nil, fmt.Errorf("convert gNMI %T type to scalar type failed: %v", e.Value, err)
				}
				stringVal := fmt.Sprint(scalarEle)
				ss[x] = stringVal
			}
		}

		i = ss
	default:
		return nil, fmt.Errorf("unsupported type %T", tv.Value)
	}
	return i, nil
}

// jsonIETFtoGNMINotification convert JSON_IETF encoded data to gNMI notifications.
// nodePath is full path of this node (from root).
func jsonIETFtoGNMINotifications(nodeJSON []byte, timeStamp int64, nodePath *gpb.Path) ([]*gpb.Notification, error) {
	nodeTemp, stat := ygotutils.NewNode(reflect.TypeOf((*ocstruct.Device)(nil)), nodePath)
	if stat.GetCode() != int32(cpb.Code_OK) {
		return nil, fmt.Errorf("cannot create empty node with path %v: %v", nodePath, stat)
	}
	nodeStruct, ok := nodeTemp.(ygot.ValidatedGoStruct)
	if !ok {
		return nil, errors.New("node is not a ValidatedGoStruct")
	}
	if err := ocstruct.Unmarshal(nodeJSON, nodeStruct); err != nil {
		return nil, fmt.Errorf("unmarshaling json data to config struct fails: %v", err)
	}
	if err := nodeStruct.Validate(); err != nil {
		return nil, err
	}
	noti, err := ygot.TogNMINotifications(nodeStruct, timeStamp, ygot.GNMINotificationsConfig{
		UsePathElem:    true,
		PathElemPrefix: nodePath.GetElem(),
	})
	if err != nil {
		return nil, fmt.Errorf("error in serializzing GoStruct to notiffications: %v", err)
	}
	return noti, nil
}

func gNMIToPrometheusMetrics(noti *gpb.Notification) ([]prometheus.Metric, error) {
	if noti == nil {
		return nil, errors.New("no gNMI telemetry in input parameters")
	}
	metrics := []prometheus.Metric{}

	prefixPath := ""
	prefixLabels := make(map[string]string)
	if noti.GetPrefix() != nil {
		prefixPath, prefixLabels = gNMIPathtoString(noti.GetPrefix())
	}
	if len(prefixPath) > 0 {
		prefixPath = prefixPath + ":"
	}
	for _, update := range noti.GetUpdate() {
		metricName, labels := gNMIPathtoString(update.GetPath())
		metricName = prefixPath + metricName
		labelKeys := []string{}
		labelValues := []string{}
		for k, v := range prefixLabels {
			labelKeys = append(labelKeys, k)
			labelValues = append(labelValues, v)
		}
		for k, v := range labels {
			labelKeys = append(labelKeys, k)
			labelValues = append(labelValues, v)
		}

		switch update.Val.Value.(type) {
		case *gpb.TypedValue_StringVal,
			*gpb.TypedValue_DecimalVal,
			*gpb.TypedValue_FloatVal,
			*gpb.TypedValue_IntVal,
			*gpb.TypedValue_UintVal,
			*gpb.TypedValue_BoolVal:
			scalarVal, err := typedValueToScalar(update.Val)
			if err != nil {
				return nil, fmt.Errorf("failed value type conversion: %v", err)
			}
			switch scalarVal.(type) {
			case string:
				// Value in string type will be saved in a label.
				labelKeys = append(labelKeys, "metric_value")
				labelValues = append(labelValues, scalarVal.(string))
				metricDesc := prometheus.NewDesc(
					metricName,
					"string type gNMI metric",
					labelKeys,
					nil,
				)
				metrics = append(metrics, prometheus.MustNewConstMetric(
					metricDesc,
					prometheus.UntypedValue,
					0,
					labelValues...,
				))
			case float32:
				metricValue := scalarVal.(float32)
				metricDesc := prometheus.NewDesc(
					metricName,
					"float32 type gNMI metric",
					labelKeys,
					nil,
				)
				metrics = append(metrics, prometheus.MustNewConstMetric(
					metricDesc,
					prometheus.GaugeValue,
					float64(metricValue),
					labelValues...,
				))
			case int64:
				metricValue := scalarVal.(int64)
				var maxFloat64, minFloat64 int64
				maxFloat64 = 2 << 52
				minFloat64 = -2 << 52
				if metricValue < minFloat64 || metricValue > maxFloat64 {
					log.Warning("Lose precision in converting int64 to float64")
				}
				metricDesc := prometheus.NewDesc(
					metricName,
					"int64 type gNMI metric",
					labelKeys,
					nil,
				)
				metrics = append(metrics, prometheus.MustNewConstMetric(
					metricDesc,
					prometheus.GaugeValue,
					float64(metricValue),
					labelValues...,
				))
			case bool:
				metricDesc := prometheus.NewDesc(
					metricName,
					"bool type gNMI metric",
					labelKeys,
					nil,
				)
				boolValue := scalarVal.(bool)
				var metricValue float64
				if boolValue {
					metricValue = 1
				} else {
					metricValue = 0
				}
				metrics = append(metrics, prometheus.MustNewConstMetric(
					metricDesc,
					prometheus.GaugeValue,
					metricValue,
					labelValues...,
				))
			case uint64:
				metricValue := scalarVal.(uint64)
				var maxFloat64 uint64
				maxFloat64 = 2 << 52
				if metricValue > maxFloat64 {
					log.Warning("Lose precision in converting uint64 to float64")
				}
				metricDesc := prometheus.NewDesc(
					metricName,
					"uint64 type gNMI metric",
					labelKeys,
					nil,
				)
				metrics = append(metrics, prometheus.MustNewConstMetric(
					metricDesc,
					prometheus.GaugeValue,
					float64(metricValue),
					labelValues...,
				))
			default:
				log.Error("Unexpected type, doesn't included in gNMI supported types")
			}
		case *gpb.TypedValue_LeaflistVal:
			scalarVal, err := typedValueToScalar(update.Val)
			if err != nil {
				return nil, fmt.Errorf("failed value type conversion: %v", err)
			}
			intfArray := scalarVal.([]interface{})
			if len(intfArray) == 0 {
				continue
			}
			labelKeys = append(labelKeys, ArrayIndexLabel)
			switch intfArray[0].(type) {
			case string:
				for idx, intf := range intfArray {
					currentLabelValues := append(labelValues, fmt.Sprint(idx))
					currentLabelKeys := append(labelKeys, "metric_value")
					currentLabelValues = append(currentLabelValues, intf.(string))
					metricDesc := prometheus.NewDesc(
						metricName,
						"string array type gNMI metric",
						currentLabelKeys,
						nil,
					)
					metrics = append(metrics, prometheus.MustNewConstMetric(
						metricDesc,
						prometheus.UntypedValue,
						0,
						currentLabelValues...,
					))
				}
			case float32:
				for idx, intf := range intfArray {
					currentLabelValues := append(labelValues, fmt.Sprint(idx))
					metricValue := intf.(float32)
					metricDesc := prometheus.NewDesc(
						metricName,
						"float32 array type gNMI metric",
						labelKeys,
						nil,
					)
					metrics = append(metrics, prometheus.MustNewConstMetric(
						metricDesc,
						prometheus.GaugeValue,
						float64(metricValue),
						currentLabelValues...,
					))
				}

			case int64:
				for idx, intf := range intfArray {
					currentLabelValues := append(labelValues, fmt.Sprint(idx))
					metricValue := intf.(int64)
					var maxFloat64, minFloat64 int64
					maxFloat64 = 2 << 52
					minFloat64 = -2 << 52
					if metricValue < minFloat64 || metricValue > maxFloat64 {
						log.Warning("Lose precision in converting int64 to float64")
					}
					metricDesc := prometheus.NewDesc(
						metricName,
						"int64 array type gNMI metric",
						labelKeys,
						nil,
					)
					metrics = append(metrics, prometheus.MustNewConstMetric(
						metricDesc,
						prometheus.GaugeValue,
						float64(metricValue),
						currentLabelValues...,
					))
				}
			case bool:
				for idx, intf := range intfArray {
					currentLabelValues := append(labelValues, fmt.Sprint(idx))
					boolValue := intf.(bool)
					var metricValue float64
					if boolValue {
						metricValue = 1
					} else {
						metricValue = 0
					}
					metricDesc := prometheus.NewDesc(
						metricName,
						"bool array type gNMI metric",
						labelKeys,
						nil,
					)
					metrics = append(metrics, prometheus.MustNewConstMetric(
						metricDesc,
						prometheus.GaugeValue,
						float64(metricValue),
						currentLabelValues...,
					))
				}
			case uint64:
				for idx, intf := range intfArray {
					currentLabelValues := append(labelValues, fmt.Sprint(idx))
					metricValue := intf.(uint64)
					var maxFloat64 uint64
					maxFloat64 = 2 << 52
					if metricValue > maxFloat64 {
						log.Warning("Lose precision in converting uint64 to float64")
					}
					metricDesc := prometheus.NewDesc(
						metricName,
						"uint64 array type gNMI metric",
						labelKeys,
						nil,
					)
					metrics = append(metrics, prometheus.MustNewConstMetric(
						metricDesc,
						prometheus.GaugeValue,
						float64(metricValue),
						currentLabelValues...,
					))
				}
			}
		case *gpb.TypedValue_JsonIetfVal:
			fullPath := update.GetPath()
			if fullPath.GetElem() != nil && noti.GetPrefix().GetElem() != nil {
				fullPath.Elem = append(fullPath.GetElem(), noti.GetPrefix().GetElem()...)
			}
			jsonIETFNoti, err := jsonIETFtoGNMINotifications(update.GetVal().GetJsonIetfVal(), noti.GetTimestamp(), fullPath)
			if err != nil {
				return nil, fmt.Errorf("Failed marshal JSON IETF into go struct: %v", err)
			}
			for _, ietfNoti := range jsonIETFNoti {
				jsonIETFMetrics, err := gNMIToPrometheusMetrics(ietfNoti)
				if err != nil {
					return nil, fmt.Errorf("failed convert gNMI notification to prometheus metrics: %v", err)
				}
				metrics = append(metrics, jsonIETFMetrics...)
			}
		case *gpb.TypedValue_JsonVal:
			var jsonValue interface{}
			if err := json.Unmarshal(update.GetVal().GetJsonVal(), &jsonValue); err != nil {
				return nil, fmt.Errorf("failed unmarshal json bolb data: %v", err)
			}
			jsonNoti := &gpb.Notification{
				Timestamp: noti.GetTimestamp(),
				Prefix:    noti.GetPrefix(),
				Update:    []*gpb.Update{},
			}
			if jsonValueMap, ok := jsonValue.(map[string]interface{}); ok {
				for k, v := range jsonValueMap {
					scalarVal, err := value.FromScalar(v)
					if err != nil {
						return nil, fmt.Errorf("failed convert json type data to gnmi type: %v", err)
					}
					nodePathElems := append([]*gpb.PathElem{}, update.GetPath().GetElem()...)
					nodePathElems = append(nodePathElems, &gpb.PathElem{Name: k})
					newUpdate := &gpb.Update{
						Path: &gpb.Path{Elem: nodePathElems},
						Val:  scalarVal,
					}
					jsonNoti.Update = append(jsonNoti.Update, newUpdate)
				}
			} else {
				scalarVal, err := value.FromScalar(jsonValue)
				if err != nil {
					return nil, fmt.Errorf("failed convert json type data to gnmi type: %v", err)
				}
				newUpdate := &gpb.Update{
					Path: update.GetPath(),
					Val:  scalarVal,
				}
				jsonNoti.Update = append(jsonNoti.Update, newUpdate)
			}
			jsonMetrics, err := gNMIToPrometheusMetrics(jsonNoti)
			if err != nil {
				return nil, fmt.Errorf("failed convert JSON data to Murdock type data: %v", err)
			}
			metrics = append(metrics, jsonMetrics...)
		}
	}
	return metrics, nil
}

// Describe export description for each fixed metrics. All YANG model's leaf node
// metrics are dynamic, they shouldn't be described here.
func (collector *APStateCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc(
		"GNMI_Prometheus_Collector",
		"This metric will not be collected, it only used to initial the collector",
		nil,
		nil,
	)
}

// Collect will read current metric value from TargetState
func (collector *APStateCollector) Collect(ch chan<- prometheus.Metric) {

	currentState := collector.state.currentState()
	metrics, err := gNMIToPrometheusMetrics(currentState)
	if err != nil {
		log.Errorf("Error in converting telemetry to metrics: %v", err)
	}

	for _, m := range metrics {
		ch <- m
	}
}
