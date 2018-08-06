package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	dto "github.com/prometheus/client_model/go"
	"github.com/golang/protobuf/proto"
	"sort"
)

type CustomSummary interface {
	prometheus.Metric
	prometheus.Collector
	UpdateContent(sum float64, count uint64, content map[float64]float64)
}

type customSummary struct {
	selfCollector

	mtx sync.Mutex // Protects every other moving part.

	// 描述区域
	desc *prometheus.Desc

	// 标签区域
	labelPairs []*dto.LabelPair

	// 数据区域
	objectives map[float64]float64
	sum        float64
	cnt        uint64
}

func (cs *customSummary) UpdateContent(sum float64, count uint64, content map[float64]float64) {
	cs.mtx.Lock()
	defer cs.mtx.Unlock()
	cs.sum = sum
	cs.cnt = count
	cs.objectives = content
}

func (cs *customSummary) Desc() *prometheus.Desc {
	return cs.desc
}

func (cs *customSummary) Write(metric *dto.Metric) error {
	sum := &dto.Summary{}
	qs := make([]*dto.Quantile, 0, len(cs.objectives))

	cs.mtx.Lock()

	for rank, value := range cs.objectives {
		qs = append(qs, &dto.Quantile{
			Quantile: proto.Float64(rank),
			Value:    proto.Float64(value),
		})
	}
	cs.mtx.Unlock()

	sum.SampleCount = proto.Uint64(cs.cnt)
	sum.SampleSum = proto.Float64(cs.sum)
	sort.Sort(sortAbleQs(qs))
	sum.Quantile = qs

	metric.Summary = sum
	metric.Label = cs.labelPairs
	return nil
}

func NewCustomSummary(opts prometheus.SummaryOpts) CustomSummary {
	labels := make([]*dto.LabelPair, len(opts.ConstLabels))
	i := 0
	for k, v := range opts.ConstLabels {
		labels[i] = &dto.LabelPair{
			Name:  proto.String(k),
			Value: proto.String(v),
		}
		i++
	}
	cs := &customSummary{
		mtx:        sync.Mutex{},
		desc:       prometheus.NewDesc(prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name), opts.Help, nil, opts.ConstLabels),
		labelPairs: labels,
		objectives: opts.Objectives,
		sum:        0,
		cnt:        0,
	}
	cs.init(cs)
	return cs
}
