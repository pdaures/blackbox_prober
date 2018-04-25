package pingers

import (
	"fmt"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

const urlTag = "url"
const hostTag = "host"

// MetricMaker creates metrics to be reported later on
type MetricMaker interface {
	MakeMetric(name string)
}

type Reporter struct {
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

func (r *Reporter) ReportLatency(latency float64, url *url.URL) {
	r.withLabelValues(r.latency, url).Set(latency)
}

func (r *Reporter) ReportSize(size int, url *url.URL) {
	r.withLabelValues(r.size, url).Set(float64(size))
}

func (r *Reporter) ReportHttpStatus(status int, url *url.URL) {
	r.withLabelValues(r.httpStatus, url).Set(float64(status))
}

func (r *Reporter) ReportSuccess(success bool, metricName string, url *url.URL) error {
	metric, ok := r.otherMetrics[metricName]
	if !ok {
		return fmt.Errorf("metric %s unknown", metricName)
	}
	successValue := 0
	if success {
		successValue = 1
	}
	r.withLabelValues(metric, url).Set(float64(successValue))
	return nil
}

func (r *Reporter) withLabelValues(g *prometheus.GaugeVec, url *url.URL) prometheus.Gauge {
	urlStr := url.String()
	hostname := url.Hostname()

	values := []string{}
	for _, tagName := range r.tagNames {
		if tagName == urlTag {
			values = append(values, urlStr)
		} else if tagName == hostTag {
			values = append(values, hostname)
		} else {
			values = append(values, r.tags[tagName])
		}
	}
	return g.WithLabelValues(values...)
}

// MakeMetric implements MetricMaker
func (r *Reporter) MakeMetric(name string) {
	if _, ok := r.otherMetrics[name]; ok {
		return
	}
	metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: r.namespace,
		Name:      name,
		Help:      name,
	}, r.tagNames)
	r.otherMetrics[name] = metric
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
