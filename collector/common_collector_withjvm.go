package collector

var jvmConfig = Properties{
	"JVM": {
		"java.lang:type=OperatingSystem": {
			&Property{"MaxFileDescriptorCount", TypeGauge, "最大文件描述合计"},
			&Property{"OpenFileDescriptorCount", TypeGauge, "打开文件描述合计"},
			&Property{"SystemCpuLoad", TypeGauge, "系统负载"},
			&Property{"ProcessCpuLoad", TypeGauge, "CPU负载"},
		}},
}

func NewJvmCollector(hostPort string) *CommonCollector {
	return NewCommonCollector(hostPort, jvmConfig, nil)
}

func NewCommonCollectorWithJvm(hostPort string, config Properties, labels map[string]string) *CommonCollector {
	return NewCommonCollector(hostPort, jvmConfig.Append(config), labels)
}
