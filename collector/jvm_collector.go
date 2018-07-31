package collector

import "github.com/prometheus/client_golang/prometheus"

type JvmMetrics struct {
	host       string
	collectors map[string]prometheus.Gauge
}

func (jc *JvmMetrics) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range jc.collectors {
		v.Describe(ch)
	}
}

func (jc *JvmMetrics) Collect(ch chan<- prometheus.Metric) {
	for _, v := range jc.collectors {
		v.Set(124564)
		v.Collect(ch)
		//todo: update metrics
	}
}

func NewJvmMetrics(host string) *JvmMetrics {
	const jvmMetrics = "Hadoop:service=NameNode,name=JvmMetrics"
	const path = "/jxm?qry=" + jvmMetrics
	var metrics = []string{
		"MemNonHeapUsedM",
		"MemNonHeapCommittedM",
		"MemNonHeapMaxM",
		"MemHeapUsedM",
		"MemHeapCommittedM",
		"MemHeapMaxM",
		//todo: more
	}
	result := &JvmMetrics{
		host:       host,
		collectors: make(map[string]prometheus.Gauge),
	}
	for _, v := range metrics {
		result.collectors[v] = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "jvm",
				Name:      v,
				Help:      "HelpOf" + v,
			})
	}
	return result
}
