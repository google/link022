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

package main

import (
	"context"
	"flag"
	"net/http"

	log "github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	targetAddr = flag.String("target_addr", "localhost:10161", "The target address in the format of host:port")
	targetName = flag.String("target_name", "hostname.com", "The target name is used to verify the hostname returned by TLS handshake")
	listenAddr = flag.String("listen_addr", "localhost:8080", "The address to listen on for HTTP requests.")
)

func main() {
	flag.Parse()
	targetState := &TargetState{}
	ctx := context.Background()
	go monitoringAPStats(ctx, *targetAddr, *targetName, targetState)
	gnmiExport := newAPStateCollector(targetState)
	prometheus.MustRegister(gnmiExport)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
