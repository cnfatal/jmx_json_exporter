package collector

import (
	"github.com/fatalc/jmx_json_exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strings"
)

const jmxEndpoint = "/jmx"

type CommonCollector struct {
	hostname string
	// todo: 增加基础jvm监控数据
	config     map[string][]string
	Collectors map[string]prometheus.Collector
}

func (bc *CommonCollector) Describe(ch chan<- *prometheus.Desc) {
	log.Printf("Describe: %s", bc.hostname)
	for _, v := range bc.Collectors {
		v.Describe(ch)
	}
}

func (bc *CommonCollector) Collect(ch chan<- prometheus.Metric) {
	log.Printf("Collect: %s", bc.hostname)
	beans, err := utils.JmxJsonBeansParse(utils.Get("http://" + bc.hostname + jmxEndpoint))
	if err != nil {
		log.Printf("Collect 未收集到数据")
		return
	}
	for k, v := range bc.Collectors {
		vars := strings.Split(k, "#")
		vGauge, ok := v.(prometheus.Gauge)
		if ok {
			data := beans[vars[0]].Content[vars[1]]
			switch data.(type) {
			case float64:
				vGauge.Set(data.(float64))
			case []interface{}:
				vGauge.Set(float64(len(data.([]interface{}))))
			}
		}
		v.Collect(ch)
	}
}

func NewBeansCollector(host string, namespace string, config map[string][]string) *CommonCollector {
	beansCollector := &CommonCollector{
		hostname:   host,
		config:     config,
		Collectors: make(map[string]prometheus.Collector),
	}
	for k, v := range config {
		for _, v2 := range v {
			beansCollector.Collectors[k+"#"+v2] = prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   namespace,
					Subsystem:   strings.Replace(strings.Split(host, ":")[0], ".", "", -1),
					Name:        v2,
					Help:        "HelpOf" + v2,
					ConstLabels: nil,
				})
		}
	}
	return beansCollector
}
