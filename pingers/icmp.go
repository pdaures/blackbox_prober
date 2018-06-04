package pingers

import (
	"log"
	"os/exec"
	"strconv"
	"time"
)

func pingerICMP(addr string, reporter MetricReporter, c *Rule) error {
	start := time.Now()
	err := exec.Command("ping", "-n", "-c", "1", "-W", strconv.Itoa(c.Timeout), addr).Run()
	if err != nil {
		log.Printf("Couldn't ping %s: %v\n", addr, err)
		reporter.ReportSuccess(false, c.MetricName, hostLabel(addr, c.tags))
		return err
	}
	reporter.ReportLatency(time.Since(start).Seconds(), hostLabel(addr, c.tags))
	reporter.ReportSuccess(true, c.MetricName, hostLabel(addr, c.tags))
	return nil
}

func hostLabel(addr string, others map[string]string) map[string]string {
	return pingerLabels(addr, addr, others)
}
