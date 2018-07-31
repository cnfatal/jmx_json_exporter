package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"github.com/fatalc/jmx_json_exporter/utils"
	"log"
)

const jmxEndpoint  = "/jmx"

type CommonCollector struct {
	hostname string
	config map[string][]string
	Collectors map[string]prometheus.Collector
}

func (bc *CommonCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range bc.Collectors {
		v.Describe(ch)
	}
}

func (bc *CommonCollector) Collect(ch chan<- prometheus.Metric) {
	body,err := utils.Get("http://" + bc.hostname + jmxEndpoint)
	if err != nil {
		log.Print("获取数据出错:" + err.Error())
	}
	beans := utils.JmxJsonBeansParse(body)
	for k, v := range bc.Collectors {
		vars := strings.Split(k,"#")
		vGauge,ok:=v.(prometheus.Gauge)
		if ok {
			data,ok := beans[vars[0]].Content[vars[1]].(float64)
			if !ok {
				continue
			}
			vGauge.Set(data)
		}
		v.Collect(ch)
	}
}

func NewBeansCollector(host string,config map[string][]string) *CommonCollector {
	beansCollector := &CommonCollector{
		hostname:   host,
		config:     config,
		Collectors: make(map[string]prometheus.Collector),
	}
	for k, v := range config {
		for _, v2 := range v {
			beansCollector.Collectors[k + "#" +v2] = prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   "host_" + strings.Replace(strings.Replace(host,".","_",-1),":","_",-1),
					Subsystem:   strings.Split(k,":")[0],
					Name:        v2,
					Help:        "HelpOf" + v2,
					ConstLabels: nil,
				})
		}
	}
	return beansCollector
}

