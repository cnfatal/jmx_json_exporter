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
	up         Gauge
	collectors map[string]interface{ Collector }
}

func (bc *CommonCollector) Describe(ch chan<- *Desc) {
	for _, v := range bc.collectors {
		v.Describe(ch)
	}
}

func (bc *CommonCollector) Collect(ch chan<- Metric) {
	beans, err := JmxJsonBeansParse(utils.Get("http://" + bc.hostname + jmxEndpoint))
	if err != nil {
		// 网络问题或服务器宕机引发获取不到监控数据判定为 down
		bc.up.Set(0)
		bc.up.Collect(ch)
		log.Printf("can't collect metrics maybe server is down")
		return
	} else {
		bc.up.(Gauge).Set(1)
		bc.up.Collect(ch)
	}
	for k, v := range bc.collectors {
		domain, name := DecodePropertyKey(k)
		switch v.(type) {
		case Gauge:
			v.(Gauge).Set(beans[domain].Content[name].(float64))
		case CustomSummary:
			v.(CustomSummary).UpdateContent(generateCustomSummaryContent(NameRegexp(name), beans[domain]))
		default:
			log.Printf("unsupport type %v", v)
		}
		v.Collect(ch)
	}
}

func NewCommonCollector(hostPort string, config Properties, labels map[string]string) *CommonCollector {
	log.Printf("initial Common Collector To -> %s", hostPort)
	if labels == nil {
		labels = map[string]string{model.InstanceLabel: strings.Split(hostPort, ":")[0]}
	} else {
		labels[model.InstanceLabel] = strings.Split(hostPort, ":")[0]
	}
	beans, err := JmxJsonBeansParse(utils.Get(httpProtocol + hostPort + jmxEndpoint))
	if err != nil {
		log.Fatal(err.Error())
	}
	// 增加 up 在线状态检测
	up := NewGauge(GaugeOpts{
		Name:        "up",
		Help:        "在线状态检测",
		ConstLabels: labels,
	})
	return &CommonCollector{hostPort, config, up, generateCollector(config, beans, labels)}
}

func generateCollector(config Properties, beans map[string]*JmxBean, labels Labels) map[string]interface{ Collector } {
	result := make(map[string]interface{ Collector })
	for namespace, properties := range config {
		for domain, items := range properties {
			for beanName, bean := range beans {
				//todo:regexp compile , domain regexp beanName
				if string(domain) == beanName {
					for _, item := range items {
						//infer if exist, if not then ignore
						if !existProperty(item, bean) {
							log.Printf("namespace:%s , %s can't found", namespace, item.NameRegexp)
							continue
						}
						switch item.DataType {
						case TypeGauge:
							result[EncodePropertyKey(beanName, item.NameRegexp)] = NewGauge(GaugeOpts{
								Namespace:   string(namespace),
								Subsystem:   inferSubSystemName(bean),
								Name:        string(item.NameRegexp),
								Help:        item.Help,
								ConstLabels: labels,
							})
						case TypeSummary:
							_, _, content := generateCustomSummaryContent(item.NameRegexp, bean)
							result[EncodePropertyKey(beanName, item.NameRegexp)] = NewCustomSummary(SummaryOpts{
								Namespace:   string(namespace),
								Subsystem:   inferSubSystemName(bean),
								Name:        string(item.NameRegexp),
								Help:        item.Help,
								ConstLabels: labels,
								Objectives:  content,
							})
						default:
							log.Printf("namespace:%s , unsupport type %s of %s", namespace, item.DataType, item.NameRegexp)
						}
					}
				}
			}
		}

	}
	return result
}

func inferSubSystemName(bean *JmxBean) string {
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
	content = make(map[float64]float64)
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

func existProperty(property *Property, bean *JmxBean) (exist bool) {
	s := string(property.NameRegexp)
	switch property.DataType {
	case TypeGauge:
		_, exist = bean.Content[s]
	case TypeSummary:
		for k := range bean.Content {
			if strings.Contains(k, s) {
				return true
			}
		}
	default:
		return false
	}
	return exist
}
