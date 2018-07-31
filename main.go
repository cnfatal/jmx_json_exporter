package main

import (
	"flag"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
		"github.com/prometheus/client_golang/prometheus"

		"github.com/fatalc/jmx_json_exporter/collector"
)

var (
	from = flag.String("from", "localhost:80/jmx", "The URL of \"/jmx\" json resources ")
	port = flag.String("out", ":8080", "The port of \"/metrics\"  output endpoint")
	path = flag.String("path", "/metrics", "Path of output endpoint")
)

func init() {
	log.Printf("initalizing")
}

func main() {
	flag.Parse()
	prometheus.MustRegister(collector.NewHadoopCollector([]string{*from}))
	http.Handle(*path, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
				<head><title>Jmx Json Exporter</title></head>
            	<body>
            		<h1>jmx json Exporter</h1>
            		<p><a href='` + *path + `'>Metrics</a></p>
            	</body>
			</html>`))
	})
	log.Printf("server listing at %v", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}
