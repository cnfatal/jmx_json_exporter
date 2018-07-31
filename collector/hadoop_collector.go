package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"github.com/fatalc/jmx_json_exporter/utils"
	"encoding/json"
	"strings"
	)

type HadoopMasterCollector struct {
	hostname   string
	jvm        *JvmMetrics
	collectors map[string]prometheus.Gauge
}

func (mc *HadoopMasterCollector) Collect(ch chan<- prometheus.Metric) {
	mc.jvm.Collect(ch)
	body, err := utils.Get(mc.hostname)
	if err != nil {
		log.Fatal(err)
	}
	beans := utils.JmxJsonBeansParse(body)
	for k, v := range mc.collectors {
		vars := strings.Split(k, "#")
		v.Set(beans[vars[0]].Content[vars[1]].(float64))
		v.Collect(ch)
	}
}

//func (mc *HadoopMasterCollector) sendData(name string, data float64) {
//	collector, exist := mc.collectors[name]
//	if !exist {
//		collector = prometheus.NewGauge(
//			prometheus.GaugeOpts{
//				Namespace:   mc.hostname,
//				Name:        name,
//				Help:        "Help_of_" + name,
//				ConstLabels: nil,
//			})
//		mc.collectors[name] = collector
//	}
//	collector.Set(data)
//}

func (mc *HadoopMasterCollector) Describe(ch chan<- *prometheus.Desc) {
	mc.jvm.Describe(ch)
	for _, v := range mc.collectors {
		v.Describe(ch)
	}
}

func NewHadoopMaterCollector(host string) *HadoopMasterCollector {
	const nameNodeInfo = "Hadoop:service=NameNode,name=NameNodeInfo"
	quotas := map[string][]string{
		nameNodeInfo: {
			"Total",
			"Used",
			"Free",
			"NonDfsUsedSpace",
		},
	}
	masterCollector := &HadoopMasterCollector{
		hostname:   host,
		jvm:        NewJvmMetrics(host),
		collectors: make(map[string]prometheus.Gauge),
	}

	for k, v := range quotas {
		for _, v2 := range v {
			masterCollector.collectors[k+"#"+v2] = prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "master",
					//Subsystem:   k,
					Name:        v2,
					Help:        "HelpOf" + v2,
					ConstLabels: nil,
				})
		}
	}
	return masterCollector
}

type HadoopWorkerCollector struct {
	hostname string
	jvm      *JvmMetrics
}

func (wc *HadoopWorkerCollector) Collect(ch chan<- prometheus.Metric) {
	//panic("implement me")
}

func (wc *HadoopWorkerCollector) Describe(ch chan<- *prometheus.Desc) {
	//panic("implement me")
}


type HadoopCollector struct {
	MasterHost        []string
	nodeHosts         []string
	masterCollectors  map[string]*HadoopMasterCollector
	workersCollectors map[string]*HadoopWorkerCollector
}

func (hc *HadoopCollector) Describe(ch chan<- *prometheus.Desc) {
	for k, v := range hc.masterCollectors {
		log.Printf("Describe of %s", k)
		v.Describe(ch)
	}
	//for k, v := range hc.workersCollectors {
	//	log.Printf("Describe of %s", k)
	//	v.Describe(ch)
	//}
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
	mcs := make(map[string]*HadoopMasterCollector, len(masterHosts))
	for _, v := range masterHosts {
		mcs[v] = NewHadoopMaterCollector(v)
	}
	wcs := make(map[string]*HadoopWorkerCollector, len(nodeHosts))
	for _, v := range nodeHosts {
		wcs[v] = &HadoopWorkerCollector{
			hostname: v,
		}
	}
	return &HadoopCollector{
		MasterHost:        masterHosts,
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
