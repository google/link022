package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openconfig/gnmi/value"

	log "github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/google/gnxi/utils/credentials"
	"github.com/google/gnxi/utils/xpath"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	apStatsExportingDelay = 15 * time.Second
	timeOut               = 10 * time.Second
	statusPath            = "/"
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
	prefixPath := ""
	prefixLabels := make(map[string]string)
	if currentState.GetPrefix() != nil {
		prefixPath, prefixLabels = gNMIPathtoString(currentState.GetPrefix())
	}
	for _, update := range currentState.GetUpdate() {
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

		updateValue, err := value.ToScalar(update.GetVal())
		if err != nil {
			log.Errorf("Error converting gNMI TypeValue to scalar type: %v", err)
			continue
		}
		switch updateValue.(type) {
		case string:
			// Value in string type will be saved in a label.
			labelKeys = append(labelKeys, "metric_value")
			labelValues = append(labelValues, updateValue.(string))
			metricDesc := prometheus.NewDesc(
				metricName,
				"string type gNMI metric",
				labelKeys,
				nil,
			)
			ch <- prometheus.MustNewConstMetric(
				metricDesc,
				prometheus.UntypedValue,
				0,
				labelValues...,
			)
		case float32:
			metricValue := updateValue.(float32)
			metricDesc := prometheus.NewDesc(
				metricName,
				"float32 type gNMI metric",
				labelKeys,
				nil,
			)
			ch <- prometheus.MustNewConstMetric(
				metricDesc,
				prometheus.GaugeValue,
				float64(metricValue),
				labelValues...,
			)
		case int64:
			metricValue := updateValue.(int64)
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
			ch <- prometheus.MustNewConstMetric(
				metricDesc,
				prometheus.GaugeValue,
				float64(metricValue),
				labelValues...,
			)
		case bool:
			metricDesc := prometheus.NewDesc(
				metricName,
				"bool type gNMI metric",
				labelKeys,
				nil,
			)
			boolValue := updateValue.(bool)
			var metricValue float64
			if boolValue {
				metricValue = 1
			} else {
				metricValue = 0
			}
			ch <- prometheus.MustNewConstMetric(
				metricDesc,
				prometheus.GaugeValue,
				metricValue,
				labelValues...,
			)
		case uint64:
			metricValue := updateValue.(uint64)
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
			ch <- prometheus.MustNewConstMetric(
				metricDesc,
				prometheus.GaugeValue,
				float64(metricValue),
				labelValues...,
			)
		case []interface{}:
			// All elements in this slice will be saved in labels.
			for idx, intf := range updateValue.([]interface{}) {
				labelKey := fmt.Sprintf("metric_value_%d", idx)
				labelValue := fmt.Sprint(intf)
				labelKeys = append(labelKeys, labelKey)
				labelValues = append(labelValues, labelValue)
			}
			metricDesc := prometheus.NewDesc(
				metricName,
				"Array type gNMI metric",
				labelKeys,
				nil,
			)
			ch <- prometheus.MustNewConstMetric(
				metricDesc,
				prometheus.UntypedValue,
				0,
				labelValues...,
			)
		case []byte:
			log.Info("Receive bytes type metric. Discard it because it's not aim for Prometheus")
		default:
			log.Error("Unknown type, doesn't include in gNMI supported types")
			continue
		}
	}

}
