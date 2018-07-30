package collector

import "github.com/prometheus/client_golang/prometheus"

type JVMCollector struct {
	collectors prometheus.Collector
}

