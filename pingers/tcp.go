package pingers

import (
	"log"
	"net"
	"strings"
	"time"
)

func pingerTCP(addr string, reporter MetricReporter, c *Rule) error {
	timeoutDuration := time.Second * time.Duration(c.Timeout)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeoutDuration)

	if err != nil {
		log.Printf("Couldn't connect to %s: %s", addr, err)
		reporter.ReportSuccess(false, c.MetricName, addrLabel(addr, c.tags))
		return err
	}
	defer conn.Close()
	reporter.ReportLatency(time.Since(start).Seconds(), addrLabel(addr, c.tags))
	reporter.ReportSuccess(true, c.MetricName, addrLabel(addr, c.tags))
	return nil
}

func addrLabel(addr string, others map[string]string) map[string]string {
	hostPort := strings.SplitN(addr, ":", 1)
	return pingerLabels(addr, hostPort[0], others)
}
