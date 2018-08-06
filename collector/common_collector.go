package collector

import (
	"github.com/fatalc/jmx_json_exporter/utils"
	. "github.com/prometheus/client_golang/prometheus"
	"log"
	"strings"
	"github.com/prometheus/common/model"
)

const httpProtocol = "http://"
const jmxEndpoint = "/jmx"

type CommonCollector struct {
	hostname   string
	config     Properties
	collectors map[string]interface{ Collector }
}

func (bc *CommonCollector) Describe(ch chan<- *Desc) {
	log.Printf("Describe: %s", bc.hostname)
	for _, v := range bc.collectors {
		v.Describe(ch)
	}
}

func (bc *CommonCollector) Collect(ch chan<- Metric) {
	log.Printf("Collect: %s", bc.hostname)
	beans, err := JmxJsonBeansParse(utils.Get("http://" + bc.hostname + jmxEndpoint))
	if err != nil {
		log.Printf("Collect 未收集到数据")
		return
	}
	for k, v := range bc.collectors {
		domain, name := DecodePropertyKey(k)
		switch v.(type) {
		case Gauge:
			v.(Gauge).Set(beans[domain].Content[name].(float64))
		case CustomSummary:
			v.(CustomSummary).UpdateContent(generateCustomSummaryContent(NameRegexp(name), beans[name]))
		default:
			log.Printf("unsupport type %v", v)
		}
		v.Collect(ch)
	}
}

func NewCommonCollector(hostPort string, namespace string, config Properties, labels map[string]string) *CommonCollector {
	if labels == nil {
		labels = map[string]string{model.InstanceLabel: strings.Split(hostPort, ":")[0]}
	} else {
		labels[model.InstanceLabel] = strings.Split(hostPort, ":")[0]
	}
	commonCollector := &CommonCollector{
		hostname:   hostPort,
		config:     config,
		collectors: make(map[string]interface{ Collector }),
	}
	beans, err := JmxJsonBeansParse(utils.Get(httpProtocol + hostPort + jmxEndpoint))
	if err != nil {
		log.Fatal(err.Error())
	}
	for domain, items := range config {
		for name, bean := range beans {
			//todo:
			if string(domain) == name {
				for _, item := range items {
					switch item.DataType {
					case TypeGauge:
						commonCollector.collectors[EncodePropertyKey(domain, item.NameRegexp)] = NewGauge(GaugeOpts{
							Namespace:   namespace,
							Subsystem:   getSubSystem(bean),
							Name:        string(item.NameRegexp),
							Help:        item.Help,
							ConstLabels: labels,
						})
					case TypeSummary:
						_, _, content := generateCustomSummaryContent(item.NameRegexp, bean)
						commonCollector.collectors[EncodePropertyKey(domain, item.NameRegexp)] = NewCustomSummary(SummaryOpts{
							Namespace:   namespace,
							Subsystem:   getSubSystem(bean),
							Name:        string(item.NameRegexp),
							Help:        item.Help,
							ConstLabels: labels,
							Objectives:  content,
						})
					default:
						log.Printf("unsupport type %s", item.DataType)
					}
				}
			}
		}
	}
	return commonCollector
}

func getSubSystem(bean *JmxBean) string {
	labels := bean.Labels
	value, ok := labels["name"]
	if ok {
		return value
	}
	value, ok = labels["type"]
	if ok {
		return value
	}
	return "UndefinedSubSystem"
}

func generateCustomSummaryContent(summaryName NameRegexp, bean *JmxBean) (sum float64, count uint64, content map[float64]float64) {
	for name, value := range bean.Content {
		if strings.Contains(name, string(summaryName)) {
			switch strings.Split(name, "_")[1] {
			case "num_ops":
				count = value.(uint64)
			case "25th":
				content[0.25] = value.(float64)
			case "median":
				content[0.5] = value.(float64)
			case "75ht":
				content[0.5] = value.(float64)
			case "90th":
				content[0.5] = value.(float64)
			case "95th":
				content[0.5] = value.(float64)
			case "99.9th":
				content[0.99] = value.(float64)
				sum = value.(float64)
			}
		}
	}
	return
}
