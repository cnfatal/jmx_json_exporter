package main

import (
	"encoding/json"
	"flag"
	"github.com/fatalc/jmx_json_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"net/http"
)

const emptyJson = "{}"

var (
	from       = flag.String("from", "localhost:8080", "The \"/jmx\"endpoint's host:port ")
	port       = flag.String("port", "9200", "The port of \"/metrics\"  output endpoint(for prometheus)")
	path       = flag.String("path", "/metrics", "The path of output endpoint")
	config     = flag.String("config", emptyJson, "Json type config string")
	configFile = flag.String("config-file", "./config.json", "Json type config file path")
)

func main() {
	flag.Parse()
	commonCollectorWithJvm := collector.NewCommonCollectorWithJvm(*from, analyseConfig(*config, *configFile), nil)
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

func analyseConfig(jsonString string, filePath string) collector.Properties {
	var configBytes []byte
	if jsonString == emptyJson {
		bytes, e := ioutil.ReadFile(filePath)
		if e != nil {
			log.Fatal(e.Error())
		}
		configBytes = bytes
	} else {
		configBytes = []byte(jsonString)
	}
	var config collector.Properties
	err := json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return config
}
