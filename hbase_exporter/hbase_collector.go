package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

var config = map[string][]string{

}

type HbaseCollector struct {
	hosts      []string
	collectors map[string]map[string]prometheus.Gauge
}

func (hc *HbaseCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, collectors := range hc.collectors {
		for _, collector := range collectors {
			collector.Describe(ch)
		}
	}
}

func (hc *HbaseCollector) Collect(ch chan<- prometheus.Metric) {
	for _, v := range hc.collectors {
		for _, v2 := range v {
			v2.Collect(ch)
		}
	}
}

func NewHbaseCollector(masterHosts []string) *HbaseCollector {
	collectors := make(map[string]map[string]prometheus.Gauge)
	for _, host := range masterHosts {
		collectors[host] = make(map[string]prometheus.Gauge)
		for beanName, values := range config {
			for _, value := range values {
				collectors[host][beanName+"#"+value] = prometheus.NewGauge(
					prometheus.GaugeOpts{
						Namespace:   "hbase",
						Subsystem:   strings.Replace(host, ":", "", -1),
						//todo: 未完成
						Name:        "",
						Help:        "",
						ConstLabels: nil,
					})
			}
		}
	}
	result := &HbaseCollector{
		hosts:      masterHosts,
		collectors: collectors,
	}
	return result
}
