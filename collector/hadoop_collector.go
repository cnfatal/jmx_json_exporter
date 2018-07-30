package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
)

var jmxEndpoint = "/jmx"

type HadoopMasterCollector struct {
	hostname string
	//todo: 具体监控指标项
}

func (mc *HadoopMasterCollector) Collect(chan<- prometheus.Metric) {
	panic("implement me")
}

func (mc *HadoopMasterCollector) Describe(chan<- *prometheus.Desc) {
	panic("implement me")
}

type HadoopWorkerCollector struct {
	hostname string
	//todo: 具体监控指标项
}

func (wc *HadoopWorkerCollector) Collect(chan<- prometheus.Metric) {
	panic("implement me")
}

func (wc *HadoopWorkerCollector) Describe(chan<- *prometheus.Desc) {
	panic("implement me")
}

type HadoopCollector struct {
	MasterHost        []string
	nodeHosts         []string
	masterCollectors  map[string]HadoopMasterCollector
	workersCollectors map[string]HadoopWorkerCollector
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

func NewHadoopCollector(masterHost []string) *HadoopCollector {
	//nodes := getNodeHosts(masterHost)
	return &HadoopCollector{
		MasterHost:        masterHost,
		nodeHosts:         getNodeHosts(masterHost),
		masterCollectors:  nil,
		workersCollectors: nil,
	}
}

func getNodeHosts(master []string) []string {
	//todo: 解析主节点路径  获取子节点监控地址
	return make([]string, 0)
}
