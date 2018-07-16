package prometheusexporter

import (
	"context"
	"flag"
	"net/http"

	log "github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	targetAddr = flag.String("target-address", "localhost:10161", "The target address in the format of host:port")
	targetName = flag.String("target_name", "hostname.com", "The target name use to verify the hostname returned by TLS handshake")
	listenAddr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
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
