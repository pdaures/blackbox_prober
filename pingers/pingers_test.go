package pingers

import "github.com/prometheus/client_golang/prometheus"

var (
	metrics = Metrics{
		Up: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "up",
			Help:      "1 if url is reachable, 0 if not",
		}, []string{"url"}),
		Latency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "latency_seconds",
			Help:      "Latency of request for url",
		}, []string{"url"}),
		Size: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "size_bytes",
			Help:      "Size of request for url",
		}, []string{"url"}),
		ExpireTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "cert_expire_timestamp",
			Help:      "Certificate expiry date in seconds since epoch.",
		}, []string{"url"}),

		StatusCode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "response_code",
			Help:      "HTTP response code.",
		}, []string{"url"}),
	}
)
