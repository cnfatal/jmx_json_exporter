package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type selfCollector struct {
	self prometheus.Metric
}

func (c *selfCollector) init(self prometheus.Metric) {
	c.self = self
}

func (c *selfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.self.Desc()
}

func (c *selfCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- c.self
}

type sortAbleQs []*dto.Quantile

func (s sortAbleQs) Len() int {
	return len(s)
}

func (s sortAbleQs) Less(i, j int) bool {
	return s[i].GetQuantile() < s[j].GetQuantile()
}

func (s sortAbleQs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

