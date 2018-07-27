package main

import (
	"flag"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"io/ioutil"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"

	"fmt"
)

var (
	in  = flag.String("in", "localhost:80/jmx", "The URL of \"/jmx\" json resources ")
	out = flag.String("out", ":8080", "The port of \"/metrics\"  output endpoint")
)

var(
	gaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "GAUGEVECNAME",
			Help:        "GAUGEVECHELP",
			ConstLabels: nil,
		},
		[]string{"service","group"},
	)

	gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "GAUGENAME",
			Help:        "GAUGEHELP",
			ConstLabels: nil,
		},
	)

	summaryVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "DEFAUILTSUMMARY",
			Help:        "-------------",
			ConstLabels: nil,
			MaxAge:      0,
			AgeBuckets:  0,
			BufCap:      0,
		},
		[]string{"service"},
	)
	histogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "DEFAULTHISTOGRAM",
			Help:        "------------",
			ConstLabels: nil,
			Buckets:     nil,
		},
	)
)

func jmxJsonParse(source []byte) []map[string]interface{} {
	jmx := make(map[string]interface{})
	json.Unmarshal(source, &jmx)
	for _, v := range jmx {
		vs := v.([]interface{})
		for i:=0;i<len(vs);i++ {
			parseMetrics(vs[i])
		}
	}
	return nil
}

func parseMetrics(mbean interface{})  {
	m := mbean.(map[string]interface{})
	name := m["name"]
	fmt.Printf("%v\n", name)
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
	summaryVec.WithLabelValues("testiiiiiiii").Observe(54612)
	prometheus.MustRegister(summaryVec)
	prometheus.MustRegister(histogram)
	gaugeVec.WithLabelValues("GAUGEEEEEE","GDS").Set(1150)
	//gaugeVec.WithLabelValues("GE2").Set(10230)
	prometheus.MustRegister(gaugeVec)
	gauge.Set(45546)
	prometheus.MustRegister(gauge)
}

func main() {
	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		bytes, err := getJmx(*in)
		if err != nil {
			log.Fatal(err.Error())
		}
		jmxJsonParse(bytes)

		//writer.Write()
	})
	log.Printf("server listing at %v", *out)
	log.Fatal(http.ListenAndServe(*out, nil))
}
