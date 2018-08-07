package jvm_collector

import (
	. "github.com/prometheus/client_golang/prometheus"
	. "github.com/fatalc/jmx_json_exporter/collector"
)

var jvmConfig = Properties{
	"JVM": {
		"java.lang:type=OperatingSystem": {
			&Property{"MaxFileDescriptorCount", TypeCounter, "最大文件描述合计"},
			&Property{"OpenFileDescriptorCount", TypeCounter, "打开文件描述合计"},
			&Property{"SystemCpuLoad", TypeGauge, "系统负载"},
			&Property{"ProcessCpuLoad", TypeGauge, "CPU负载"},
		}},
}

type JvmCollector interface {
	Collector
}

func NewJvmCollector(hostPort string) JvmCollector {
	return NewCommonCollector(hostPort, jvmConfig, nil)
}

func NewWithJvmCollector(hostPort string, config Properties, labels map[string]string) JvmCollector {
	return NewCommonCollector(hostPort, jvmConfig.Append(config), labels)
}
