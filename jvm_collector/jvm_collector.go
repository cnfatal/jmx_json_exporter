package jvm_collector

import (
	. "github.com/prometheus/client_golang/prometheus"
	. "github.com/fatalc/jmx_json_exporter/collector"
)

const (
	nameSpace = "JVM"
)

var config = Properties{
	"java.lang:type=OperatingSystem": {
		&PropertiesItem{"MaxFileDescriptorCount", TypeCounter, "最大文件描述合计"},
		&PropertiesItem{"OpenFileDescriptorCount", TypeCounter, "打开文件描述合计"},
		&PropertiesItem{"SystemCpuLoad", TypeGauge, "系统负载"},
		&PropertiesItem{"ProcessCpuLoad", TypeGauge, "CPU负载"},
	},
}

type JvmCollector interface {
	Collector
}

type jvmCollector struct {
	labels   map[string]string
	config   Properties
	hostPort string
	*CommonCollector
}

func NewJvmCollector(hostPort string) JvmCollector {
	return &jvmCollector{
		labels:          nil,
		config:          config,
		hostPort:        hostPort,
		CommonCollector: NewCommonCollector(hostPort,nameSpace,config,nil),
	}
}
