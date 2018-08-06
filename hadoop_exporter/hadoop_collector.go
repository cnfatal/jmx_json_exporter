package main

import (
	"encoding/json"
	. "github.com/fatalc/jmx_json_exporter/collector"
	"github.com/fatalc/jmx_json_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strings"
	"github.com/fatalc/jmx_json_exporter/jvm_collector"
)

const namespace = "hadoop"

// 配置需要监控的数据项 todo：现在仅支持两层嵌套,待改进
var (
	masterConfig = Properties{
		"Hadoop:service=NameNode,name=NameNodeInfo": {
			&PropertiesItem{"Total", TypeGauge, "磁盘总计",},
			&PropertiesItem{"Used", TypeGauge, "已使用",},
			&PropertiesItem{"Free", TypeGauge, "剩余",},
			&PropertiesItem{"NonDfsUsedSpace", TypeGauge, "非DFS使用空间",},
		},
	}
	workerConfig = Properties{
		"Hadoop:service=DataNode,name=JvmMetrics": {
			&PropertiesItem{"MemNonHeapCommittedM", TypeGauge, "已使用内存",		},
			&PropertiesItem{"LogWarn", TypeGauge, "日志警告",		},
		},
	}
)

type HadoopCollector struct {
	masterHosts       map[string]string
	nodeHosts         map[string]string
	masterCollectors  map[string]*CommonCollector
	workersCollectors map[string]*CommonCollector
}

func (hc *HadoopCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range hc.masterCollectors {
		v.Describe(ch)
	}
	for _, v := range hc.workersCollectors {
		v.Describe(ch)
	}
}

//Collect implements the prometheus.Collector interface. 该接口调用来更新数据
func (hc *HadoopCollector) Collect(ch chan<- prometheus.Metric) {
	for _, v := range hc.masterCollectors {
		v.Collect(ch)
	}
	for _, v := range hc.workersCollectors {
		v.Collect(ch)
	}
}

func NewHadoopCollector(masterHosts map[string]string) *HadoopCollector {
	nodeHosts := getNodeHosts(masterHosts)
	mcs := make(map[string]*CommonCollector, len(masterHosts))
	for _, masterHostPort := range masterHosts {
		// todo: 使用hostname代替ip
		mcs[masterHostPort] = jvm_collector.NewWithJvmCollector(masterHostPort, namespace, masterConfig)
	}
	wcs := make(map[string]*CommonCollector, len(nodeHosts))
	for host, v := range nodeHosts {
		wcs[v] = NewCommonCollector(host, namespace, workerConfig)
	}
	return &HadoopCollector{
		masterHosts:       masterHosts,
		nodeHosts:         nodeHosts,
		masterCollectors:  mcs,
		workersCollectors: wcs,
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
