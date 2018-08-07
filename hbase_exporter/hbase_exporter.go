package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"github.com/fatalc/jmx_json_exporter/utils"
	"strings"
	"strconv"
	. "github.com/fatalc/jmx_json_exporter/collector"
)

var (
	from = flag.String("from", "localhost:80", "The host of Zookeeper Server ")
	port = flag.String("port", "8080", "The port of \"/metrics\"  output endpoint")
	path = flag.String("path", "/metrics", "Path of output endpoint")
)

var (
	hbaseConfig = Properties{
		"hbase": {
			"Hadoop:service=HBase,name=Master,sub=Quotas": {
				&Property{"SnapshotObserverSizeComputationTime",TypeSummary,"快照"},
				&Property{"numNamespaceInQuotaViolation", TypeGauge, "help"},
				&Property{"SnapshotQuotaObserverChoreTime", TypeSummary, "help"},
				&Property{"numRegionSizeReports", TypeGauge, "help"},
				&Property{"numTablesInQuotaViolation", TypeGauge, "help"},
				&Property{"QuotaObserverChoreTime", TypeSummary, "help"},
			},
			"Hadoop:service=HBase,name=Master,sub=Server": {
				&Property{"masterActiveTime", TypeGauge, "help"},
				&Property{"averageLoad", TypeGauge, "help"},
				&Property{"clusterRequests", TypeGauge, "help"},
			},
			"Hadoop:service=HBase,name=Master,sub=FileSystem":{
				&Property{"HlogSplitTime",TypeSummary,"hlog"},
				&Property{"MetaHlogSplitTime",TypeSummary,"hlog"},
			},
			"Hadoop:service=HBase,name=RegionServer,sub=IO": {
				&Property{"FsPReadTime",TypeSummary,"fs读取时间"},
				&Property{"FsReadTime",TypeSummary,"fs读取时间"},
				&Property{"FsWriteTime",TypeSummary,"fs写入时间"},
			},
			"Hadoop:service=HBase,name=RegionServer,sub=Server": {
				&Property{"regionServerStartTime", TypeGauge, "regionServer启动时间"},
				&Property{"percentFilesLocal", TypeGauge, "help"},
				&Property{"percentFilesLocalSecondaryRegions", TypeGauge, "help"},
			},
			"Hadoop:service=HBase,name=RegionServer,sub=Memory": {
				&Property{"memStoreSize", TypeGauge, "内存总计"},
				&Property{"unblockedFlushGauge", TypeGauge, "help"},
			},
		},
	}
)

type HbaseCollector struct {
	masterHosts []string
	collectors  map[string]*CommonCollector
}

func (hc *HbaseCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, commonCollector := range hc.collectors {
		commonCollector.Describe(ch)
	}
}

func (hc *HbaseCollector) Collect(ch chan<- prometheus.Metric) {
	for _, v := range hc.collectors {
		v.Collect(ch)
	}

}

func getRegionServers(masterHosts []string) *map[string]string {
	const jmxEndpoint = "/jmx"
	const httpProtocol = "http://"
	const serverStatus = "Hadoop:service=HBase,name=Master,sub=Server"
	const liveRegions = "tag.liveRegionServers"
	const deadRegions = "tag.deadRegionServers"
	const emptyString = ""
	result := make(map[string]string)
	for _, host := range masterHosts {
		beans, err := JmxJsonBeansParse(utils.Get(httpProtocol + host + jmxEndpoint))
		if err != nil {
			continue
		}
		lives := beans[serverStatus].Content[liveRegions].(string)
		deads := beans[serverStatus].Content[deadRegions].(string)
		if lives != emptyString {
			for _, regionStr := range strings.Split(lives, ";") {
				regionArr := strings.Split(regionStr, ",")
				// 推测 regionServer 的http 界面地址
				result[regionArr[0]+":"+inferJmxPort(regionArr[1])] = regionArr[0] + ":" + inferJmxPort(regionArr[1])
			}
		}
		if deads != emptyString {
			for _, regionStr := range strings.Split(deads, ";") {
				regionArr := strings.Split(regionStr, ",")
				result[regionArr[0]+":"+inferJmxPort(regionArr[1])] = regionArr[0] + ":" + inferJmxPort(regionArr[1])
			}
		}
	}
	log.Printf("Analysised Regionservers: %v", result)
	return &result
}

func inferJmxPort(port string) string {
	const inferNum = 10
	intPort, err := strconv.Atoi(port)
	if err != nil {
		return port
	}
	inferPort := intPort + inferNum
	log.Printf("infered port %d is %d", port, inferPort)
	return strconv.Itoa(inferPort)
}

func NewHbaseCollector(masterHosts []string) *HbaseCollector {
	result := &HbaseCollector{
		masterHosts: masterHosts,
		collectors:  make(map[string]*CommonCollector),
	}
	for _, master := range masterHosts {
		result.collectors[master] = NewCommonCollectorWithJvm(master, hbaseConfig, nil)
	}
	for _, region := range *getRegionServers(masterHosts) {
		result.collectors[region] = NewCommonCollectorWithJvm(region, hbaseConfig, nil)
	}
	return result
}

func main() {
	flag.Parse()
	prometheus.MustRegister(NewHbaseCollector([]string{*from}))
	http.Handle(*path, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
				<head><title>Hbase Exporter</title></head>
           	<body>
           		<h1>Hbase Exporter</h1>
				<p><a href='` + *path + `'>Metrics</a></p>
           	</body>
			</html>`))
	})
	listenAddress := ":" + *port
	log.Printf("server listing at %v", ":"+*port)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
