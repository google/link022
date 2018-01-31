/* Copyright 2017 Google Inc.

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

// Package controller contains methods that related to AP controller.
package controller

import (
	"errors"
	"time"

	devctx "github.com/google/link022/agent/context"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	log "github.com/golang/glog"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	hostnameHeader       = "hostname"
	gnmiAddrHeader       = "gnmi_address"
	resyncRequiredHeader = "resync_required"

	heartbeatInterval = 15 * time.Second
)

// ReportAPInfo reports the current AP information to the assigned controller.
// It returns error if failures detected.
func ReportAPInfo(syncRequried bool) error {
	deviceConfig := devctx.GetDeviceConfig()
	if deviceConfig.ControllerAddr == "" {
		return errors.New("no controller assigned, unable to report AP information")
	}

	// Creating a gNMI client.
	conn, err := grpc.Dial(deviceConfig.ControllerAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	getRequest := &pb.GetRequest{
		Path: []*pb.Path{},
	}

	// Add AP's gNMI information to the request header.
	headerMap := map[string]string{
		hostnameHeader: deviceConfig.Hostname,
		gnmiAddrHeader: deviceConfig.GNMIServerAddr,
	}
	if syncRequried {
		headerMap[resyncRequiredHeader] = "True"
	}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(headerMap))

	cli := pb.NewGNMIClient(conn)
	if _, err := cli.Get(ctx, getRequest); err != nil {
		return err
	}
	return nil
}

// Connect sends heartbeat messages to the assigned controller periodically.
// The first (received) message contains a sync request to fetch the latest AP configuration.
// This function never returns, should be run in the background.
func Connect() {
	syncRequired := true
	for {
		if err := ReportAPInfo(syncRequired); err != nil {
			log.Errorf("Cannot connect to the controller, retry in %s. Error: %v.", heartbeatInterval, err)
			// Disconncetion detected, do a re-sync.
			syncRequired = true
		} else if syncRequired {
			log.Info("Controller connected and received the sync request.")
			syncRequired = false
		}
		time.Sleep(heartbeatInterval)
	}
}
