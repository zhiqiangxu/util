package monitor

// this module registers metrics and pprof to http.DefaultServeMux

import (
	"net/http"
	// enable pprof
	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	http.Handle("/metrics", promhttp.Handler())
}
