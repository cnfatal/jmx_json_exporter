package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"github.com/fatalc/jmx_json_exporter/utils"
	"encoding/json"
	"strings"
)

// 配置需要监控的数据项 todo：现在仅支持两层嵌套,待改进
var (
	masterConfig = map[string][]string{
		"Hadoop:service=NameNode,name=NameNodeInfo": {
			"Total",
			"Used",
			"Free",
			"NonDfsUsedSpace",
		},
	}
	workerConfig = map[string][]string{
		"Hadoop:service=DataNode,name=JvmMetrics": {
			"MemNonHeapCommittedM",
			"LogWarn",
		},
	}
)

type HadoopCollector struct {
	masterHosts       []string
	nodeHosts         []string
	masterCollectors  map[string]*CommonCollector
	workersCollectors map[string]*CommonCollector
}

func (hc *HadoopCollector) Describe(ch chan<- *prometheus.Desc) {
	for k, v := range hc.masterCollectors {
		log.Printf("Describe of %s", k)
		v.Describe(ch)
	}
	for k, v := range hc.workersCollectors {
		log.Printf("Describe of %s", k)
		v.Describe(ch)
	}
}

//Collect implements the prometheus.Collector interface. 该接口调用来更新数据
func (hc *HadoopCollector) Collect(ch chan<- prometheus.Metric) {
	for k, v := range hc.masterCollectors {
		log.Printf("Collect of %s", k)
		v.Collect(ch)
	}
	for k, v := range hc.workersCollectors {
		log.Printf("Collect of %s", k)
		v.Collect(ch)
	}
}

func NewHadoopCollector(masterHosts []string) *HadoopCollector {
	nodeHosts := getNodeHosts(masterHosts)
	mcs := make(map[string]*CommonCollector, len(masterHosts))
	for _, v := range masterHosts {
		mcs[v] = NewBeansCollector(v, masterConfig)
	}
	wcs := make(map[string]*CommonCollector, len(nodeHosts))
	for _, v := range nodeHosts {
		wcs[v] = NewBeansCollector(v, workerConfig)
	}
	return &HadoopCollector{
		masterHosts:       masterHosts,
		nodeHosts:         nodeHosts,
		masterCollectors:  mcs,
		workersCollectors: wcs,
	}
}

func getNodeHosts(masterHosts []string) []string {
	const protocol = "http://"
	const nameNodeInfo = "Hadoop:service=NameNode,name=NameNodeInfo"
	const path = "/jmx?qry=" + nameNodeInfo
	const liveNodesName = "LiveNodes"
	const infoKey = "infoAddr"

	nodeUrls := make(map[string]bool)
	for _, v := range masterHosts {
		body, err := utils.Get(protocol + v + path)
		if err != nil {
			log.Fatal(err)
		}
		liveNodes := utils.JmxJsonBeansParse(body)[nameNodeInfo].Content[liveNodesName].(string)
		nodesJson := make(map[string]interface{})
		json.Unmarshal([]byte(strings.Trim(liveNodes, "/")), &nodesJson)
		for _, v := range nodesJson {
			nodeUrl := v.(map[string]interface{})[infoKey].(string)
			nodeUrls[nodeUrl] = true
		}
	}
	result := make([]string, len(nodeUrls))
	i := 0
	for k := range nodeUrls {
		result[i] = k
		i ++
	}
	return result
}
