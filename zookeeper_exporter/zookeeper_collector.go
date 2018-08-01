package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/fatalc/jmx_json_exporter/utils"
	"encoding/json"
	"strings"
	"log"
)

var counter = 0

const (
	endpoint = "/commands"
)

var config = map[string][]string{
	"dirs": {
		"datadir_size",
		"logdir_size",
	},
	"monitor": {
		"avg_latency",
		"max_latency",
		"min_latency",
		"packets_received",
		"packets_sent",
		"num_alive_connections",
		"znode_count",
		"watch_count",
		"ephemerals_count",
		"approximate_data_size",
		"open_file_descriptor_count",
	},
	"serverStats": {
		//"server_stats",
		"node_count",
	},
	"stats": {
		//"server_stats",
		"connections",
	},
	"watch_summary": {
		"num_connections",
		"num_paths",
		"num_total_watches",
	},
}

type zooKeeperCollector struct {
	hosts      []string
	collectors map[string]map[string]prometheus.Gauge
}

func (zc *zooKeeperCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range zc.collectors {
		for _, v2 := range v {
			v2.Describe(ch)
		}
	}
}

func getCommand(command string, host string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal(utils.Get("http://"+host+endpoint+"/"+command), &result)
	return result
}

func (zc *zooKeeperCollector) Collect(ch chan<- prometheus.Metric) {
	counter ++
	log.Printf("Do Collect(), %d times", counter)
	for _, host := range zc.hosts {
		for command, values := range config {
			commandReturns := getCommand(command, host)
			for _, value := range values {
				//collector, ok := zc.collectors[host][command+"$"+value].(prometheus.Gauge)
				collector := zc.collectors[host][command+"#"+value]
				if collector == nil {
					continue
				}
				//if !ok {
				//	continue
				//} else {
				switch commandReturns[value].(type) {
				case float64:
					collector.Set(commandReturns[value].(float64))
					break
				case []interface{}:
					collector.Set(float64(len(commandReturns[value].([]interface{}))))
					break
				}
				collector.Collect(ch)
				//}
			}
		}
	}
}

func NewZookeeperCollector(hosts []string) *zooKeeperCollector {
	collectors := make(map[string]map[string]prometheus.Gauge)
	for _, host := range hosts {
		collectors[host] = make(map[string]prometheus.Gauge)
	}
	result := &zooKeeperCollector{
		hosts:      hosts,
		collectors: collectors,
	}
	for _, host := range hosts {
		for command, values := range config {
			for _, value := range values {
				result.collectors[host][command+"#"+value] = prometheus.NewGauge(
					prometheus.GaugeOpts{
						Namespace:   "zookeeper",
						Subsystem:   strings.Replace(strings.Replace(host, ":", "", -1), ".", "", -1),
						Name:        command + "_" + value,
						Help:        "Help",
						ConstLabels: nil,
					})
			}
		}
	}
	return result
}
