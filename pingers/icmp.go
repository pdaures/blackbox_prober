package pingers

import (
	"log"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func pingerICMP(url *url.URL, reporter *Reporter, c *Rule) error {
	hostPort := strings.Split(url.Host, ":")
	start := time.Now()
	err := exec.Command("ping", "-n", "-c", "1", "-W", strconv.Itoa(c.Timeout), hostPort[0]).Run()
	if err != nil {
		log.Printf("Couldn't ping %s: %s", url, err)
		return reporter.ReportSuccess(false, c.MetricName, url)
	}
	reporter.ReportLatency(time.Since(start).Seconds(), url)
	return reporter.ReportSuccess(true, c.MetricName, url)
}
