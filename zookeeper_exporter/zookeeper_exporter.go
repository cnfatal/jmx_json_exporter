package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strings"
)

const zooHttpPort = ":8080"

var (
	from = flag.String("from", "localhost", "The qurom of Zookeeper ")
	port = flag.String("port", "8080", "The port of \"/metrics\"  output endpoint")
	path = flag.String("path", "/metrics", "Path of output endpoint")
)

func init() {
	log.Printf("initalizing")
}

func main() {
	flag.Parse()

	zookeepers := strings.Split(*from, ",")
	vars := make([]string, len(zookeepers))
	for index, zookeeper := range zookeepers {
		vars[index] = zookeeper + zooHttpPort
	}
	prometheus.MustRegister(NewZookeeperCollector(vars))
	http.Handle(*path, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
				<head><title>Zookeeper Exporter</title></head>
           	<body>
           		<h1>Zookeeper Exporter</h1>
				<p><a href='` + *path + `'>Metrics</a></p>
           	</body>
			</html>`))
	})
	listenAddress := ":" + *port
	log.Printf("server listing at %v", ":8080")
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
