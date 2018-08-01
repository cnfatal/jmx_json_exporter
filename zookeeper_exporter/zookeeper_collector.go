package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/fatalc/jmx_json_exporter/utils"
	"encoding/json"
	"strings"
	"log"
)

const (
	endpoint    = "/commands"
	dirs        = "dirs"
	monitor     = "monitor"
	serverStats = "server_stats"
	stats       = "stats"
)

var config = map[string][]string{
	dirs: {
		"datadir_size",
		"logdir_size",
	},
	monitor: {
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
	serverStats: {
		//"server_stats",
		"node_count",
	},
	stats: {
		//"server_stats",
		"connections",
	},
}

type zooKeeperCollector struct {
	hosts      []string
	collectors map[string]map[string]prometheus.Collector
}

func (zc *zooKeeperCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range zc.collectors {
		for _, v2 := range v {
			v2.Describe(ch)
		}
	}
}

func getCommand(command string, host string) map[string]interface{} {
	log.Printf("Query:%s/%s",host,command)
	result := make(map[string]interface{})
	json.Unmarshal(utils.Get("http://"+host+endpoint+"/"+command), &result)
	return result
}

func (zc *zooKeeperCollector) Collect(ch chan<- prometheus.Metric) {
	for _, host := range zc.hosts {
		for command, values := range config {
			commandReturns := getCommand(command, host)
			for _, value := range values {
				collector, ok := zc.collectors[host][command+"$"+value].(prometheus.Gauge)
				if !ok {
					continue
				} else {
					switch commandReturns[value].(type) {
					case float64:
						collector.Set(commandReturns[value].(float64))
					case []interface{}:
						collector.Set(float64(len(commandReturns[value].([]interface{}))))
					}
					collector.Collect(ch)
				}
			}
		}
	}
}

func NewZookeeperCollector(hosts []string) *zooKeeperCollector {
	collectors := make(map[string]map[string]prometheus.Collector)
	for _, host := range hosts {
		collectors[host] = make(map[string]prometheus.Collector)
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
						Subsystem:   "node_" + strings.Replace(host,":","",-1),
						Name:        command + "_" + value,
						Help:        "Help",
						ConstLabels: nil,
					})
			}
		}
	}
	return result
}
