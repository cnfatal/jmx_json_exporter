package jvm_collector

import (
	. "github.com/prometheus/client_golang/prometheus"
	. "github.com/fatalc/jmx_json_exporter/collector"
)

const (
	nameSpace = "JVM"
)

var jvmConfig = Properties{
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
	labels          map[string]string
	config          Properties
	hostPort        string
	commonCollector *CommonCollector
	jvmCollector    *CommonCollector
}

func (jc *jvmCollector) Collect(ch chan<- Metric) {
	jc.jvmCollector.Collect(ch)
	if jc.commonCollector != nil {
		jc.commonCollector.Collect(ch)
	}
}

func (jc *jvmCollector) Describe(ch chan<- *Desc) {
	jc.jvmCollector.Describe(ch)
	if jc.commonCollector != nil {
		jc.commonCollector.Describe(ch)
	}
}

func NewJvmCollector(hostPort string) JvmCollector {
	return &jvmCollector{
		labels:       nil,
		config:       jvmConfig,
		hostPort:     hostPort,
		jvmCollector: NewCommonCollector(hostPort, nameSpace, jvmConfig, nil),
	}
}

func NewWithJvmCollector(hostPort string, namespace string, config Properties, labels map[string]string) JvmCollector {
	return &jvmCollector{
		labels:          nil,
		config:          config,
		hostPort:        hostPort,
		jvmCollector:    NewCommonCollector(hostPort, nameSpace, jvmConfig, nil),
		commonCollector: NewCommonCollector(hostPort, namespace, config, labels),
	}
}
