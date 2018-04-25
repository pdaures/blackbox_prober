package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/pdaures/blackbox_prober/pingers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gopkg.in/yaml.v2"
)

var (
	listenAddress = flag.String("web-listen-address", ":9110", "Address to listen on for web interface and telemetry.")
	metricsPath   = flag.String("web-telemetry-path", "/metrics", "Path under which to expose metrics.")
	configPath    = flag.String("conf-path", "blackbox.yml", "Configuration file path.")

	errNoPinger = errors.New("No pinger for schema")
)

func main() {
	flag.Parse()

	fmt.Printf("Starting blackbox-exporter on %s%s, using configuration file: %s\n", *metricsPath, *listenAddress, *configPath)
	b, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c := &pingers.Configuration{}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	pingCollector, err := newCollector(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	prometheus.MustRegister(pingCollector)
	http.Handle(*metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

type collector struct {
	reporter *pingers.Reporter
	targets  []*pingers.Target
}

func newCollector(conf *pingers.Configuration) (*collector, error) {
	reporter := pingers.NewReporter(conf.Namespace, conf.Tags)
	targets, err := pingers.NewTargets(conf, reporter)
	if err != nil {
		return nil, err
	}
	return &collector{
		reporter: reporter,
		targets:  targets,
	}, nil
}

// Collect implements prometheus.Collector.
func (c collector) Collect(ch chan<- prometheus.Metric) {
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
func (c collector) Describe(ch chan<- *prometheus.Desc) {
	c.reporter.Describe(ch)
}
