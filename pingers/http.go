package pingers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func pingerHTTP(urlStr string, reporter MetricReporter, r *Rule) error {

	URL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("cannot parse url %s, %v\n", urlStr, err)
		reporter.ReportSuccess(false, r.MetricName, map[string]string{})
		return err
	}

	httpRule := r.HTTPRule
	metricName := r.MetricName
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: httpRule.Insecure},
			DisableKeepAlives: true,
		},
		Timeout: time.Second * time.Duration(r.Timeout),
	}
	start := time.Now()
	resp, err := client.Get(urlStr)
	if err != nil {
		log.Printf("Couldn't get %s: %v", urlStr, err)
		reporter.ReportSuccess(false, metricName, urlLabels(URL, r.tags))
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, httpRule.ReadMax))
	if err != nil {
		log.Printf("Couldn't read HTTP body for %s: %v", urlStr, err)
		reporter.ReportSuccess(false, metricName, urlLabels(URL, r.tags))
		return err
	}
	size := len(body)
	reporter.ReportLatency(time.Since(start).Seconds(), urlLabels(URL, r.tags))
	reporter.ReportSize(size, urlLabels(URL, r.tags))
	reporter.ReportHttpStatus(resp.StatusCode, urlLabels(URL, r.tags))

	match := matchBody(body, httpRule)
	validStatus := validStatus(resp.StatusCode, httpRule)

	ok := match && validStatus
	if ok && httpRule.PayloadExtractRule != nil {
		val, err := extractValue(body, httpRule)
		if err != nil {
			fmt.Printf("cannot extract value from HTTP response, %v\n", err)
		} else {
			reporter.ReportValue(val, httpRule.PayloadExtractRule.MetricName, urlLabels(URL, r.tags))
		}
	}
	reporter.ReportSuccess(ok, metricName, urlLabels(URL, r.tags))
	return nil
}

func extractValue(body []byte, httpRule *HTTPRule) (float64, error) {
	cmd := exec.Command("jq", httpRule.PayloadExtractRule.JQQuery)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, string(body))
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("error: cannot execute jq: %s\n", out)
		return 0, err
	}
	outStr := fmt.Sprintf("%s", out)
	outStr = strings.TrimSpace(outStr)
	val, err := strconv.ParseFloat(outStr, 64)
	if err != nil {
		fmt.Printf("error: cannot convert %s to float, %v", out, err)
	}
	return val, err
}

func matchBody(body []byte, httpRule *HTTPRule) bool {
	if httpRule.CompiledRegex != nil {
		return httpRule.CompiledRegex.Match(body)
	}
	if httpRule.BodyContentBytes != nil && len(httpRule.BodyContentBytes) > 0 {
		return bytes.Equal(bytes.TrimSpace(body), httpRule.BodyContentBytes)
	}
	return true
}

func validStatus(status int, httpRule *HTTPRule) bool {
	if httpRule.IgnoreHTTPStatus {
		return true
	}
	if httpRule.ValidHTTPStatuses != nil && len(httpRule.ValidHTTPStatuses) > 0 {
		for _, stat := range httpRule.ValidHTTPStatuses {
			if stat == status {
				return true
			}
		}
		return false
	}
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

func urlLabels(URL *url.URL, others map[string]string) map[string]string {
	return pingerLabels(URL.String(), URL.Hostname(), others)
}
