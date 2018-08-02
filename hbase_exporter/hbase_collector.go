package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/fatalc/jmx_json_exporter/collector"
	"github.com/fatalc/jmx_json_exporter/utils"
	"strings"
	"log"
	"strconv"
)

const nameSpace = "Hbase"
const jmxEndpoint = "/jmx"
const httpProtocol = "http://"

var (
	masterConfig = map[string][]string{
		"Hadoop:service=HBase,name=Master,sub=Quotas": {
			"numNamespaceInQuotaViolation",
			"SnapshotObserverSizeComputationTime_num_ops",
			"SnapshotQuotaObserverChoreTime_num_ops",
			"numRegionSizeReports",
			"numTablesInQuotaViolation",
			"QuotaObserverChoreTime_num_ops",
			"SnapshotObserverSnapshotFetchTime_num_ops",
		},
		"Hadoop:service=HBase,name=Master,sub=Server": {
			"masterActiveTime",
			"averageLoad",
			"clusterRequests",
		},
	}
	regionConfig = map[string][]string{
		"Hadoop:service=HBase,name=RegionServer,sub=IO": {
			"FsPReadTime_num_ops",
			"FsWriteTime_num_ops",
			"FsReadTime_num_ops",
			"FsPReadTime_max",
		},
		"Hadoop:service=HBase,name=RegionServer,sub=Server": {
			"regionServerStartTime",
			"percentFilesLocal",
			"percentFilesLocalSecondaryRegions",
		},
		"Hadoop:service=HBase,name=RegionServer,sub=Memory": {
			"memStoreSize",
			"unblockedFlushGauge",
		},
	}
)

type HbaseCollector struct {
	masterHosts []string
	masters     map[string]*collector.CommonCollector
	regions     map[string]*collector.CommonCollector
}

func (hc *HbaseCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, commonCollector := range hc.masters {
		commonCollector.Describe(ch)
	}
	for _, commonCollector := range hc.regions {
		commonCollector.Describe(ch)
	}
}

func (hc *HbaseCollector) Collect(ch chan<- prometheus.Metric) {
	for _, v := range hc.masters {
		v.Collect(ch)
	}
	for _, v := range hc.regions {
		v.Collect(ch)
	}

}

func getRegionServers(masterHosts []string) *map[string]string {
	const serverStatus = "Hadoop:service=HBase,name=Master,sub=Server"
	const liveRegions = "tag.liveRegionServers"
	const deadRegions = "tag.deadRegionServers"
	const emptyString = ""

	result := make(map[string]string)

	for _, host := range masterHosts {
		beans, err := utils.JmxJsonBeansParse(utils.Get(httpProtocol + host + jmxEndpoint))
		if err != nil {
			continue
		}
		lives := beans[serverStatus].Content[liveRegions].(string)
		deads := beans[serverStatus].Content[deadRegions].(string)

		if lives != emptyString {
			for _, regionStr := range strings.Split(lives, ";") {
				regionArr := strings.Split(regionStr, ",")
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
		masters:     make(map[string]*collector.CommonCollector),
		regions:     make(map[string]*collector.CommonCollector),
	}
	for _, master := range masterHosts {
		result.masters[master] = collector.NewBeansCollector(master, nameSpace, masterConfig)
	}
	for _, region := range *getRegionServers(masterHosts) {
		result.regions[region] = collector.NewBeansCollector(region, nameSpace, regionConfig)
	}
	return result
}
