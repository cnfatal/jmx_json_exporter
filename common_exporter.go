package main

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
		"github.com/prometheus/client_golang/prometheus"
	"github.com/fatalc/jmx_json_exporter/jvm_collector"
)

func init()  {
	summary := jvm_collector.NewJvmCollector("node170:9870")
	prometheus.MustRegister(summary)
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9200", nil)
}
