package collector

import "strings"

type DataType string

const (
	TypeGauge     DataType = "Gauge"
	TypeSummary   DataType = "CustomSummary"
	TypeCounter   DataType = "CustomCounter"
	TypeHistogram DataType = "CustomHistogram"
	spliteChar    string   = "^"
)

type DomainRegexp string
type NameRegexp string
type NameSpace string
type Property struct {
	NameRegexp NameRegexp
	DataType   DataType
	Help       string
}
type Properties map[NameSpace]map[DomainRegexp][]*Property

func EncodePropertyKey(domain string, name NameRegexp) string {
	return string(domain) + spliteChar + string(name)
}

func DecodePropertyKey(key string) (domain string, name string) {
	vars := strings.Split(string(key), spliteChar)
	return vars[0], vars[1]
}

func (p Properties) Append(add Properties) Properties {
	for nameSpace, value := range add {
		p[nameSpace] = value
	}
	return p
}
