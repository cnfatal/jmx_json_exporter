# Jmx Json Exporter

**Deprecated**

I'm not maintain it,beacuse it is so so so bad, you can use it just for fun 😬

Export [Dropwizard Metrics](https://github.com/dropwizard/metrics)
（also called `Yammer Metrics`）from `/jmx` endpoint to [Prometheus Metrics](https://github.com/prometheus).

It suitable for `hadoop`,`hbase`,`spark`... who using `Deopwizard Metrics` and exporter to http interface.

There are also some OOTB exporters, see: [Hadoop Exporter](/hadoop_exporter) , [Hbase Exporter](/hbase_exporter), [Zookeeper Exporter](/zookeeper_exporter)

## Getting start

```bash
make
./jmx_json_exporter --from [host]:[port]
```

Then visit http://localhost:9200 ,you can see the metrics.

## Params

|option|default|description|
|---|---|---|
|--from|localhost:8080|jmx json endpoint|
|--port|9200|output port|
|--path|/metrics|output path|
|--config|"{}"|json type  config string|
|--config-file|./config.json|config file json|

## Config

a json format config in file or via commandline is needed:
```json
{
  "namSpace":{
    "foo":[
      {"name":"foo","type":"typeGauge","help":"help-msg"},
      {"name":"bar","type":"typeCustomSummary","help":"help-msg"}
    ],
    "bar":[
      {"name":"foo","type":"typeCustomSummary","help":"help-msg"},
      {"name":"baz","type":"typeCustomSummary","help":"help-msg"}
    ]
  }  
}
```

note: The `type` filed only support `Gauge`,`CustomConter`,`CustomSummary` .

> config json support multi nameSpace

## Example
 
There is a jmx json metrics export from `http://SomeService:[port]/jmx`
```json
{
    "beans": [
        {
            "name": "java.lang:type=OperatingSystem",
            "modelerType": "sun.management.OperatingSystemImpl",
            "MaxFileDescriptorCount": 4096,
            "OpenFileDescriptorCount": 282,
            "CommittedVirtualMemorySize": 2890313728,
            "FreePhysicalMemorySize": 163299328,
            "FreeSwapSpaceSize": 4159524864,
            "ProcessCpuLoad": 0,
            "ProcessCpuTime": 1504810000000,
            "SystemCpuLoad": 0.004694835680751174,
            "TotalPhysicalMemorySize": 3973193728,
            "TotalSwapSpaceSize": 4160745472,
            "Arch": "amd64",
            "SystemLoadAverage": 0,
            "AvailableProcessors": 4,
            "Version": "3.10.0-514.el7.x86_64",
            "Name": "Linux",
            "ObjectName": "java.lang:type=OperatingSystem"
        },
        {
            "name" : "Hadoop:service=HBase,name=Master,sub=Balancer",
            "description" : "Metrics about HBase master balancer",
            "modelerType" : "Master,sub=Balancer",
            "tag.Context" : {
              "description" : "Metrics context",
              "value" : "master"
            },
            "tag.Hostname" : {
              "description" : "Local hostname",
              "value" : "node170"
            },
            "miscInvocationCount" : 5,
            "BalancerCluster_num_ops" : 1,
            "BalancerCluster_min" : 0,
            "BalancerCluster_max" : 0,
            "BalancerCluster_mean" : 0,
            "BalancerCluster_25th_percentile" : 0,
            "BalancerCluster_median" : 0,
            "BalancerCluster_75th_percentile" : 0,
            "BalancerCluster_90th_percentile" : 0,
            "BalancerCluster_95th_percentile" : 0,
            "BalancerCluster_98th_percentile" : 0,
            "BalancerCluster_99th_percentile" : 0,
            "BalancerCluster_99.9th_percentile" : 0,
            "BalancerCluster_TimeRangeCount_600000-inf" : 1
          }
    ]
}
```

The correspond config will like:

```json
{
  "someService":{
    "java.lang:type=OperatingSystem":[
      {"name":"MaxFileDescriptorCount","type":"Gauge","help":"maxFD"},
      {"name":"OpenFileDescriptorCount","type":"Gauge","help":"help-msg"}
    ]
  },
  "hbase":{
    "Hadoop:service=HBase,name=Master,sub=Balancer":[
      {"name":"BalancerCluster","type":"CustomSummary","help":"BalanceclusterSummary"}
    ]
  }
}
```

> For Summary metrics only use thr prefix,"BalancerCluster_mean" using "BalancerCluster".

or via commandline

```bash
./jmx_json_exporter --from node170:9870 --config {\"someService\":{\"java.lang:type=OperatingSystem\":[{\"name\":\"MaxFileDescriptorCount\",\"type\":\"Gauge\",\"help\":\"maxFD\"},{\"name\":\"OpenFileDescriptorCount\",\"type\":\"Gauge\",\"help\":\"help-msg\"}]}}
```

The Prometheus Metrics will be :

```text
# HELP someService_OperatingSystem_MaxFileDescriptorCount maxFD
# TYPE someService_OperatingSystem_MaxFileDescriptorCount gauge
someService_OperatingSystem_MaxFileDescriptorCount{instance="SomeService"} 4096
# HELP someService_OperatingSystem_OpenFileDescriptorCount help-msg
# TYPE someService_OperatingSystem_OpenFileDescriptorCount gauge
someService_OperatingSystem_OpenFileDescriptorCount{instance="SomeService"} 279
# HELP hbase_Master_BalancerCluster BalanceclusterSummary
# TYPE hbase_Master_BalancerCluster summary
hbase_Master_BalancerCluster{instance="192.168.34.170",quantile="0.25"} 0
hbase_Master_BalancerCluster{instance="192.168.34.170",quantile="0.5"} 0
hbase_Master_BalancerCluster{instance="192.168.34.170",quantile="0.99"} 0
hbase_Master_BalancerCluster_sum{instance="192.168.34.170"} 0
hbase_Master_BalancerCluster_count{instance="192.168.34.170"} 0
```

> The "OperatingSystem" is auto detected from string"java.lang:type=OperatingSystem", 'name' or 'type' will be used.

## Docker

```bash
make
docker run -it --rm -p 9200:9200 jmx-json-exporter --from [ip]:[port] --config {\"someService\":{\"java.lang:type=OperatingSystem\":[{\"name\":\"MaxFileDescriptorCount\",\"type\":\"Gauge\",\"help\":\"maxFD\"},{\"name\":\"OpenFileDescriptorCount\",\"type\":\"Gauge\",\"help\":\"help-msg\"}]}}
```
