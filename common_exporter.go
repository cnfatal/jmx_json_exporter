package main

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/fatalc/jmx_json_exporter/collector"
	"flag"
	"log"
)

// only for example
var customConfig = collector.Properties{
	"customComponent": {
		"nameRegexp": {
			&collector.Property{"quotaname", collector.TypeGauge, "help-msg"},
		},
	},
}

var (
	from = flag.String("from", "localhost:8080", "The \"/jmx\"endpoint's host:port ")
	port = flag.String("port", "9200", "The port of \"/metrics\"  output endpoint(for prometheus)")
	path = flag.String("path", "/metrics", "The path of output endpoint")
)

func main() {
	flag.Parse()
	commonCollectorWithJvm := collector.NewCommonCollectorWithJvm(*from, customConfig, nil)
	prometheus.MustRegister(commonCollectorWithJvm)
	http.Handle(*path, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
				<head><title>Jmx Json Exporter</title></head>
            	<body>
            		<h1>Jmx Json Exporter</h1>
            		<p><a href='` + *path + `'>Metrics</a></p>
            	</body>
			</html>`))
	})
	listenAddress := ":" + *port
	log.Printf("server listing at %v", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
