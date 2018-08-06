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
type Properties map[DomainRegexp][]*PropertiesItem
type PropertiesItem struct {
	NameRegexp NameRegexp
	DataType   DataType
	Help       string
}

func EncodePropertyKey(domain DomainRegexp, name NameRegexp) string {
	return string(domain) + spliteChar + string(name)
}

func DecodePropertyKey(key string) (domain string, name string) {
	vars := strings.Split(string(key), spliteChar)
	return vars[0], vars[1]
}
