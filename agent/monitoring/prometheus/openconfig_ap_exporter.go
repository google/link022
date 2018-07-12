package prometheusexporter

import (
	"context"
	"strings"
	"sync"
	"time"

	log "github.com/golang/glog"
	"google.golang.org/grpc"

	"github.com/google/gnxi/utils/credentials"
	"github.com/google/gnxi/utils/xpath"

	ocst "github.com/google/link022/generated/ocstruct"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	apStatsExportingDelay = 15 * time.Second
	timeOut               = 10 * time.Second
	encodingName          = "JSON_IETF"
	statusPath            = "/access-points/access-point"
)

var (
	link022ModelData = []*gpb.ModelData{{
		Name:         "office-ap",
		Organization: "Google, Inc.",
		Version:      "0.1.0",
	}}
)

// TargetState contain current state of the assigned ap device
type TargetState struct {
	state          *ocst.Device
	stateTimeStamp int64
	mutex          sync.RWMutex
}

func (s *TargetState) updateState(state *ocst.Device, timeStamp int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state = state
	s.stateTimeStamp = timeStamp
}

func (s *TargetState) currentState() (*ocst.Device, int64) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state, s.stateTimeStamp
}

func monitoringAPStats(ctx context.Context, targetAddress string, targetName string, targetState *TargetState) {
	opts := credentials.ClientCredentials(targetName)
	conn, err := grpc.Dial(targetAddress, opts...)
	if err != nil {
		log.Errorf("Dialing to %q failed: %v", targetAddress, err)
		return
	}
	defer conn.Close()

	cli := gpb.NewGNMIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	encoding, ok := gpb.Encoding_value[encodingName]
	if !ok {
		var gnmiEncodingList []string
		for _, name := range gpb.Encoding_name {
			gnmiEncodingList = append(gnmiEncodingList, name)
		}
		log.Errorf("Supported encodings: %s", strings.Join(gnmiEncodingList, ", "))
		return
	}

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
			Path:     pathList,
			Encoding: gpb.Encoding(encoding),
			UseModels: []*gpb.ModelData{{
				Name:         "office-ap",
				Organization: "Google, Inc.",
				Version:      "0.1.0",
			}},
		}

		getResponse, err := cli.Get(ctx, getRequest)
		if err != nil {
			log.Errorf("Get failed: %v", err)
		}
		for _, notif := range getResponse.GetNotification() {
			timeStamp := notif.GetTimestamp()
			for _, update := range notif.GetUpdate() {
				updatePath := update.GetPath()
				if notif.Prefix != nil {
					updatePath.Elem = append(notif.Prefix.Elem, updatePath.Elem...)
				}

				loadedAP := &ocst.Device{}
				err = ocst.Unmarshal(update.GetVal().GetJsonIetfVal(), loadedAP)
				if err != nil {
					log.Errorf("Error unmarshal JSON: %v", err)
					continue
				}
				targetState.updateState(loadedAP, timeStamp)
			}
		}
	}
}
