package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/pdaures/blackbox_prober/pingers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gopkg.in/yaml.v2"
)

var (
	listenAddress = flag.String("web.listen-address", ":9110", "Address to listen on for web interface and telemetry.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	configPath    = flag.String("conf.file-path", "blackbox.yml", "Configuration file path.")

	errURLNotAbsolute = errors.New("URL not absolute")
	errNoPinger       = errors.New("No pinger for schema")
)

type pingCollector struct {
	reporter *pingers.Reporter
	targets  []*pingers.Target
}

// newPingCollector returns a new pingCollector
func newPingCollector(conf *pingers.Configuration) (*pingCollector, error) {
	reporter := pingers.NewReporter(conf.Namespace, conf.Tags)
	targets, err := pingers.NewTargets(conf, reporter)
	if err != nil {
		return nil, err
	}
	return &pingCollector{
		reporter: reporter,
		targets:  targets,
	}, nil
}

// Collect implements prometheus.Collector.
func (c pingCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	for _, target := range c.targets {
		wg.Add(1)
		go func(target *pingers.Target) {
			defer wg.Done()
			if err := pingers.Ping(target, c.reporter); err != nil {
				fmt.Println(err)
			}
		}(target)
	}
	wg.Wait()
	c.reporter.Collect(ch)
}

// Describe implements prometheus.Collector.
func (c pingCollector) Describe(ch chan<- *prometheus.Desc) {
	c.reporter.Describe(ch)
}

func main() {
	fmt.Printf("Starting blackbox-exporter on %s%s, using configuration file: %s\n", *metricsPath, *listenAddress, *configPath)
	b, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}
	c := &pingers.Configuration{}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		panic(err)
	}
	pingCollector, err := newPingCollector(c)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(pingCollector)
	http.Handle(*metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
