package collector

import (
	"encoding/json"
	"errors"
		"strings"
)

//JmxBeans  jmx ，
// seems may like：{ "name" : "Hadoop:service=NameNode,name=JvmMetrics", ... }
// Name:Hadoop Labels:{service=NameNode,name=JVMMetrics} Content:{name={...},keys={...}}
type JmxBean struct {
	Domain  string
	Labels  map[string]string
	Content map[string]interface{}
}

//JmxJsonBeansParse can unmarshal []byte in json format
func JmxJsonBeansParse(httpBodyBytes []byte) (result map[string]*JmxBean, err error) {
	jmx := make(map[string]interface{})
	json.Unmarshal(httpBodyBytes, &jmx)
	beans, ok := jmx["beans"].([]interface{})
	if !ok {
		return nil, errors.New("can't find \"beans\" data")
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
