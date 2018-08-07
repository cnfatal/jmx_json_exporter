package main

import (
	"encoding/json"
	. "github.com/fatalc/jmx_json_exporter/collector"
	"github.com/fatalc/jmx_json_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strings"
)

const namespace = "hadoop"

// 配置需要监控的数据项
var (
	hadoopConfig = Properties{
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
)

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
		collectors[nodeHostPort] = NewCommonCollectorWithJvm(nodeHostPort, hadoopConfig,nil)
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
