package pingers

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const urlTag = "url"
const hostTag = "host"

// MetricMaker creates metrics to be reported later on
type MetricMaker interface {
	MakeMetric(name string)
}

// MetricReporter reports metrics to Prometheus
type MetricReporter interface {
	ReportLatency(latency float64, labels map[string]string)
	ReportSize(size int, labels map[string]string)
	ReportHttpStatus(status int, labels map[string]string)
	ReportSuccess(success bool, metricName string, labels map[string]string)
	ReportValue(val float64, metricName string, labels map[string]string)
}

type Reporter struct {
	mu           *sync.Mutex
	namespace    string
	tags         map[string]string
	tagNames     []string
	latency      *prometheus.GaugeVec
	size         *prometheus.GaugeVec
	httpStatus   *prometheus.GaugeVec
	otherMetrics map[string]*prometheus.GaugeVec
}

func NewReporter(namespace string, tags map[string]string) *Reporter {
	if tags == nil {
		tags = map[string]string{}
	}
	var tagNames = []string{urlTag, hostTag}
	for tagName := range tags {
		tagNames = append(tagNames, tagName)
	}
	return &Reporter{
		mu:        &sync.Mutex{},
		namespace: namespace,
		tags:      tags,
		tagNames:  tagNames,
		latency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "latency_seconds",
			Help:      "Latency of request for url",
		}, tagNames),
		size: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "size_bytes",
			Help:      "Size of request for url",
		}, tagNames),
		httpStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "response_code",
			Help:      "HTTP response code.",
		}, tagNames),
		otherMetrics: make(map[string]*prometheus.GaugeVec),
	}
}

func (r *Reporter) ReportLatency(latency float64, labels map[string]string) {
	r.size.With(labels).Set(float64(latency))
}

func (r *Reporter) ReportSize(size int, labels map[string]string) {
	r.size.With(labels).Set(float64(size))
}

func (r *Reporter) ReportHttpStatus(status int, labels map[string]string) {
	r.httpStatus.With(labels).Set(float64(status))
}

func (r *Reporter) ReportSuccess(success bool, metricName string, labels map[string]string) {
	successValue := 0
	if success {
		successValue = 1
	}
	r.ReportValue(float64(successValue), metricName, labels)
}

func (r *Reporter) ReportValue(val float64, metricName string, labels map[string]string) {
	metric := r.getMetric(metricName)
	metric.With(labels).Set(val)
}

func (r *Reporter) getMetric(name string) *prometheus.GaugeVec {
	var metric *prometheus.GaugeVec
	ok := false
	r.mu.Lock()
	if metric, ok = r.otherMetrics[name]; !ok {
		metric = r.makeMetric(name)
	}
	r.mu.Unlock()
	return metric
}

func (r *Reporter) makeMetric(name string) *prometheus.GaugeVec {
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: r.namespace,
		Name:      name,
		Help:      name,
	}, r.tagNames)
	r.otherMetrics[name] = metric
	return metric
}

// Collect implements prometheus.Collector.
func (r *Reporter) Collect(ch chan<- prometheus.Metric) {
	r.latency.Collect(ch)
	r.size.Collect(ch)
	r.httpStatus.Collect(ch)
	for _, metric := range r.otherMetrics {
		metric.Collect(ch)
	}
}

// Describe implements prometheus.Collector.
func (r *Reporter) Describe(ch chan<- *prometheus.Desc) {
	r.latency.Describe(ch)
	r.size.Describe(ch)
	r.httpStatus.Describe(ch)
	for _, metric := range r.otherMetrics {
		metric.Describe(ch)
	}
}

func pingerLabels(addr string, hostname string, others map[string]string) map[string]string {
	labels := make(map[string]string)
	for key, val := range others {
		labels[key] = val
	}
	labels[hostTag] = hostname
	labels[urlTag] = addr
	return labels
}
