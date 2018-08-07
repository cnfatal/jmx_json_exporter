package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"github.com/fatalc/jmx_json_exporter/utils"
	"encoding/json"
	"strings"
	. "github.com/fatalc/jmx_json_exporter/collector"
)

var (
	from = flag.String("from", "localhost:80", "The host of Hadoop nameNode ")
	port = flag.String("port", "8080", "The port of \"/metrics\"  output endpoint")
	path = flag.String("path", "/metrics", "Path of output endpoint")
)

// 配置需要监控的数据项
var hadoopConfig = Properties{
	"hadoop": {
		"Hadoop:service=NameNode,name=NameNodeInfo": {
			&Property{"Total", TypeGauge, "磁盘总计",},
			&Property{"Used", TypeGauge, "已使用",},
			&Property{"Free", TypeGauge, "剩余",},
			&Property{"NonDfsUsedSpace", TypeGauge, "非DFS使用空间",},
		},
		"Hadoop:service=DataNode,name=JvmMetrics": {
			&Property{"MemNonHeapCommittedM", TypeGauge, "已使用内存",},
			&Property{"LogWarn", TypeGauge, "日志警告",},
		},
	},
}

type HadoopCollector struct {
	masterHosts map[string]string
	nodeHosts   map[string]string
	collectors  map[string]*CommonCollector
}

func (hc *HadoopCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range hc.collectors {
		v.Describe(ch)
	}
}

//Collect implements the prometheus.Collector interface. 该接口调用来更新数据
func (hc *HadoopCollector) Collect(ch chan<- prometheus.Metric) {
	for _, v := range hc.collectors {
		v.Collect(ch)
	}
}

func NewHadoopCollector(masterHosts map[string]string) *HadoopCollector {
	nodeHosts := getNodeHosts(masterHosts)
	collectors := make(map[string]*CommonCollector, len(masterHosts))
	for _, masterHostPort := range masterHosts {
		collectors[masterHostPort] = NewCommonCollectorWithJvm(masterHostPort, hadoopConfig, nil)
	}
	for _, nodeHostPort := range nodeHosts {
		collectors[nodeHostPort] = NewCommonCollectorWithJvm(nodeHostPort, hadoopConfig, nil)
	}
	return &HadoopCollector{
		masterHosts: masterHosts,
		nodeHosts:   nodeHosts,
		collectors:  collectors,
	}
}

func getNodeHosts(masterHosts map[string]string) map[string]string {
	const protocol = "http://"
	const nameNodeInfo = "Hadoop:service=NameNode,name=NameNodeInfo"
	const path = "/jmx?qry=" + nameNodeInfo
	const liveNodesName = "LiveNodes"
	const infoKey = "infoAddr"

	nodeUrls := make(map[string]string)
	for _, v := range masterHosts {
		beans, err := JmxJsonBeansParse(utils.Get(protocol + v + path))
		if err != nil {
			log.Printf("%s 下未找到 Datanode 或地址无法访问")
			continue
		}
		liveNodes := beans[nameNodeInfo].Content[liveNodesName].(string)
		nodesJson := make(map[string]interface{})
		json.Unmarshal([]byte(strings.Trim(liveNodes, "/")), &nodesJson)
		for host, v := range nodesJson {
			nodeUrl := v.(map[string]interface{})[infoKey].(string)
			host = strings.Split(host, ":")[0] + ":" + strings.Split(nodeUrl, ":")[1]
			nodeUrls[host] = host
		}
	}
	return nodeUrls
}

func init() {
	log.Printf("initalizing")
}

func main() {
	flag.Parse()
	prometheus.MustRegister(NewHadoopCollector(map[string]string{*from: *from}))
	http.Handle(*path, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
				<head><title>Hadoop Exporter</title></head>
            	<body>
            		<h1>Hadoop Exporter</h1>
            		<p><a href='` + *path + `'>Metrics</a></p>
            	</body>
			</html>`))
	})
	listenAddress := ":" + *port
	log.Printf("server listing at %v", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
