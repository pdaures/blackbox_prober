package pingers

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func pingerHTTP(url *url.URL, reporter MetricReporter, r *Rule) error {
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
	resp, err := client.Get(url.String())
	if err != nil {
		log.Printf("Couldn't get %s: %v", url, err)
		err = reporter.ReportSuccess(false, metricName, url)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, httpRule.ReadMax))
	if err != nil {
		log.Printf("Couldn't read HTTP body for %s: %v", url, err)
		err = reporter.ReportSuccess(false, metricName, url)
		return err
	}
	size := len(body)
	reporter.ReportLatency(time.Since(start).Seconds(), url)
	reporter.ReportSize(size, url)
	reporter.ReportHttpStatus(resp.StatusCode, url)

	match := matchBody(body, httpRule)
	validStatus := validStatus(resp.StatusCode, httpRule)

	ok := match && validStatus
	return reporter.ReportSuccess(ok, metricName, url)
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
