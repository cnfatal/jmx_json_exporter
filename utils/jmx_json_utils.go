package utils

import (
	"encoding/json"
	"log"
	"strings"
	"errors"
)

//JmxBeans 解析后的 jmx 数据，
// 例子：{ "name" : "Hadoop:service=NameNode,name=JvmMetrics", ... }
// Name:Hadoop Labels:{service=NameNode,name=JVMMetrics} Content:{name={...},keys={...}}
type JmxBean struct {
	Domain  string
	Labels  map[string]string
	Content map[string]interface{}
}

func JmxJsonBeansParse(httpBodyBytes []byte) (result map[string]*JmxBean,err error) {
	jmx := make(map[string]interface{})
	json.Unmarshal(httpBodyBytes, &jmx)
	beans, ok := jmx["beans"].([]interface{})
	if !ok {
		return nil,errors.New("响应中未找到\"beans\" 数据")
	}
	result = make(map[string]*JmxBean, len(beans))
	for i := 0; i < len(beans); i++ {
		bean := beans[i].(map[string]interface{})
		name := bean["name"].(string)
		domain, labels := parseJmxBeanName(name)
		result[name] = &JmxBean{Domain: domain, Labels: labels, Content: bean}
	}
	return
}

func parseMetrics(name string, data interface{}, deep int, labels *map[string]string) {
	deep = deep - 1
	if deep <= 0 {
		return
	}
	switch data.(type) {
	case map[string]interface{}:
		{
			for k, v := range data.(map[string]interface{}) {
				name = name + "_" + k
				parseMetrics(name, v, deep, labels)
			}
		}
	case []interface{}:
		for _, v := range data.([]interface{}) {
			parseMetrics(name, v, deep, labels)
		}
	default:
		{
			switch data.(type) {
			case float64:

			case int:
				log.Printf("%s : %d", name, data)
			case string:
				log.Printf("string type: %s", data)
			default:
				log.Printf("unkown type %v", data)
			}
		}
	}

}

func parseJmxBeanName(name string) (domain string, properties map[string]string) {
	properties = make(map[string]string)
	var1 := strings.Split(name, ":")
	domain = var1[0]
	var2 := strings.Split(var1[1], ",")
	for _, v := range var2 {
		var3 := strings.Split(v, "=")
		properties[var3[0]] = var3[1]
	}
	return
}
