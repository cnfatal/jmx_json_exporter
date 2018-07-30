package main

import (
	"flag"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"io/ioutil"
	"github.com/prometheus/client_golang/prometheus"

	"strings"
)

var (
	from = flag.String("from", "localhost:80/jmx", "The URL of \"/jmx\" json resources ")
	port = flag.String("out", ":8080", "The port of \"/metrics\"  output endpoint")
	path = flag.String("path", "/metrics", "Path of output endpoint")
)

var (
	registered = make(map[string]prometheus.Gauge)
)

func upDateData(name string, labels *map[string]string, data float64) {
	name = strings.Replace(name, ".", "_", -1)
	gauge, exist := registered[name]
	if !exist {
		gauge = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   "",
				Subsystem:   "",
				Name:        name,
				Help:        "Metrics_of" + name,
				ConstLabels: *labels,
			},
		)
		err := prometheus.Register(gauge)
		if err != nil {
			log.Print(err.Error())
			return
		}
		registered[name] = gauge
	}
	gauge.Set(data)
}

func getJmx(url string) (bytes []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func init() {
	log.Printf("initalizing")
}

func main() {
	flag.Parse()
	http.Handle(*path, promhttp.Handler())
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		bytes, err := getJmx(*from)
		if err != nil {
			log.Fatal(err.Error())
		}
		jmxJsonParse(bytes)

		//writer.Write()
	})
	/*
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
	*/
	log.Printf("server listing at %v", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}
