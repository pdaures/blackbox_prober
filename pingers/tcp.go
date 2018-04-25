package pingers

import (
	"log"
	"net"
	"net/url"
	"time"
)

func pingerTCP(url *url.URL, reporter *Reporter, c *Rule) error {
	timeoutDuration := time.Second * time.Duration(c.Timeout)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", url.Host, timeoutDuration)
	if err != nil {
		log.Printf("Couldn't connect to %s: %s", url.Host, err)
		return reporter.ReportSuccess(false, c.MetricName, url)
	}
	defer conn.Close()
	if url.Path != "" {
		conn.SetDeadline(time.Now().Add(timeoutDuration))

		size, err := readSize(conn)
		if err != nil {
			log.Printf("Error reading from %s: %s", url, err)
			return reporter.ReportSuccess(false, c.MetricName, url)
		}
		reporter.ReportSize(size, url)
	}
	reporter.ReportLatency(time.Since(start).Seconds(), url)
	return reporter.ReportSuccess(true, c.MetricName, url)
}
