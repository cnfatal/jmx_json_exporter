package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strings"
	"github.com/prometheus/common/model"
	"encoding/json"
	"github.com/fatalc/jmx_json_exporter/utils"
)

const zooHttpPort = ":8080"

var (
	from = flag.String("from", "localhost", "The qurom of Zookeeper ")
	port = flag.String("port", "8080", "The port of \"/metrics\"  output endpoint")
	path = flag.String("path", "/metrics", "Path of output endpoint")
)

const (
	nameSpace = "zookeeper"
	endpoint  = "/commands"
	splitChar = "^"
)

type zkQuota struct {
	name string
	help string
}

var zookeeperConfig = map[string][]zkQuota{
	"dirs": {
		zkQuota{"datadir_size", "数据目录大小"},
		zkQuota{"logdir_size", "日志目录大小"},
	},
	"monitor": {
		zkQuota{"avg_latency", "平均latency"},
		zkQuota{"max_latency", "最大latency"},
		zkQuota{"min_latency", "最小latency"},
		zkQuota{"packets_received", "包-接受"},
		zkQuota{"packets_sent", "包-发送"},
		zkQuota{"num_alive_connections", "活动连接"},
		zkQuota{"znode_count", "znode计数"},
		zkQuota{"watch_count", "watch计数"},
		zkQuota{"ephemerals_count", "ephemerals计数"},
		zkQuota{"approximate_data_size", "近似数据大小"},
		zkQuota{"open_file_descriptor_count", "打开的文件描述计数"},
	},
	"serverStats": {
		zkQuota{"server_stats.packets_sent", "包-发送"},
		zkQuota{"server_stats.packets_received", "包-接受"},
		zkQuota{"node_count", "节点数"},
		zkQuota{"num_alive_client_connections", "活动连接"},
	},
	"stats": {
		//zkQuota{"connections", ""},
	},
	"watch_summary": {
		zkQuota{"num_connections", "连接数"},
		zkQuota{"num_paths", "path数"},
		zkQuota{"num_total_watches", "watch数"},
	},
}

type zooKeeperCollector struct {
	hosts      []string
	config     map[string][]zkQuota
	up         prometheus.Gauge
	collectors map[string]map[string]prometheus.Gauge
}

func (zc *zooKeeperCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range zc.collectors {
		for _, v2 := range v {
			v2.Describe(ch)
		}
	}
}

func (zc *zooKeeperCollector) Collect(ch chan<- prometheus.Metric) {
	for _, host := range zc.hosts {
		for command, quotas := range zc.config {
			returns := getCommandData(command, host)
			for _, quota := range quotas {
				collector, exist := zc.collectors[host][command+splitChar+quota.name]
				if !exist {
					continue
				}
				splits := strings.Split(quota.name, ".")
				switch len(splits) {
				case 1:
					float, ok := returns[splits[0]].(float64)
					if !ok {
						continue
					}
					collector.Set(float)
				case 2:
					var1, ok := returns[splits[0]].(map[string]interface{})
					if !ok {
						continue
					}
					var2, ok := var1[splits[1]].(float64)
					if !ok {
						continue
					}
					collector.Set(var2)
				}
				collector.Collect(ch)
			}
		}
	}
}

func NewZookeeperCollector(hosts []string) *zooKeeperCollector {
	collectors := make(map[string]map[string]prometheus.Gauge)
	for _, host := range hosts {
		// 添加 instance label
		labels := prometheus.Labels{model.InstanceLabel: host}
		collectors[host] = make(map[string]prometheus.Gauge)
		for command, quotas := range zookeeperConfig {
			for _, quota := range quotas {
				collectors[host][command+splitChar+quota.name] = prometheus.NewGauge(prometheus.GaugeOpts{
					Namespace:   nameSpace,
					Subsystem:   command,
					Name:        strings.Replace(quota.name, ".", "_", -1),
					Help:        quota.help,
					ConstLabels: labels,
				})
			}
		}
	}
	result := &zooKeeperCollector{
		hosts:      hosts,
		config:     zookeeperConfig,
		collectors: collectors,
	}
	return result
}

func getCommandData(command string, host string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal(utils.Get("http://"+host+endpoint+"/"+command), &result)
	return result
}

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
