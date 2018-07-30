package main

import (
	"flag"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"io/ioutil"
	"encoding/json"
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

func jmxJsonParse(source []byte) []map[string]interface{} {
	jmx := make(map[string]interface{})
	json.Unmarshal(source, &jmx)
	for _, var1 := range jmx {
		for _, var2 := range var1.([]interface{}) {
			main, labels := parseJmxBeanName(var2.(map[string]interface{})["name"].(string))
			parseMetrics(main, var2, 3, &labels)
		}
	}
	return nil
}

func parseMetrics(name string, data interface{}, deep int, labels *map[string]string) {
	deep = deep - 1
	if deep <= 0 {
		return
	}
	switch data.(type) {
	case map[string]interface{}:
		{
			for k, v := range data.(map[string]interface{}) {
				name = name + "_" + k
				parseMetrics(name, v, deep, labels)
			}
		}
	case []interface{}:
		for _, v := range data.([]interface{}) {
			parseMetrics(name, v, deep, labels)
		}
	default:
		{
			switch data.(type) {
			case float64:
				upDateData(name, labels, data.(float64))
			case int:
				log.Printf("%s : %d", name, data)
			case string:
				log.Printf("string type: %s", data)
			default:
				log.Printf("unkown type %v", data)
			}
		}
	}

}

func parseJmxBeanName(name string) (main string, properties map[string]string) {
	properties = make(map[string]string)
	var1 := strings.Split(name, ":")
	main = var1[0]
	var2 := strings.Split(var1[1], ",")
	for _, v := range var2 {
		var3 := strings.Split(v, "=")
		properties[var3[0]] = var3[1]
	}
	return
}

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
