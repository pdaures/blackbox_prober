package pingers

import (
	"fmt"
	"net/url"
	"regexp"
)

// DefaultTimeout is the default timeout if not specified
const DefaultTimeout = 10

// DefaultMetricName is the default exported metric name if not specified
const DefaultMetricName = "Up"

// DefaultReadMax is the default max size of body to read
const DefaultReadMax = 1e+7

// Configuration contains the rules and targets for these rules.
// This is the data structure parsed from YAML
type Configuration struct {
	Tags      map[string]string   `yaml:"tags,omitempty"` // custom tags to put in each metric
	Namespace string              `yaml:"namespace"`      // namespace added to Prometheus metric name
	Rules     map[string]*Rule    `yaml:"rules"`          // contains pinger rule, how to call and check the response
	Targets   map[string][]string `yaml:"targets"`        // mapping Rule name to URLs
}

// Rule is a definition of asserts to do on a ping.
type Rule struct {
	Type       string    `yaml:"type"`                  // tcp, http or icmp
	Timeout    int       `yaml:"timeout,omitempty"`     // timeout in seconds
	MetricName string    `yaml:"metric_name,omitempty"` // metric name used for health report, default value is Up
	HTTPRule   *HTTPRule `yaml:"http,omitempty"`        // is only required for type http
}

// HTTPRule contains the configuration for the list of http checks to do
type HTTPRule struct {
	IgnoreHTTPStatus  bool           `yaml:"ignore_http_status,omitempty"` // ignore HTTP status for health report
	ValidHTTPStatuses []int          `yaml:"statuses,omitempty"`
	BodyContentBytes  []byte         `yaml:"-"`
	BodyContent       string         `yaml:"body_content,omitempty"` // if set, the HTTP response body must be BodyContent
	BodyRegex         string         `yaml:"body_regexp,omitempty"`  // if set, the HTTP response body must match BodyRegex
	CompiledRegex     *regexp.Regexp `yaml:"-"`
	Insecure          bool           `yaml:"insecure,omitempty"`
	ReadMax           int64          `yaml:read_max,omitempty`
}

// Target is a the definition of the check to execute (which rule on which endpoint)
type Target struct {
	Name string
	URL  *url.URL
	Rule *Rule
}

func NewTargets(c *Configuration, metricMaker MetricMaker) ([]*Target, error) {
	for _, rule := range c.Rules {
		err := rule.Setup(metricMaker)
		if err != nil {
			return nil, err
		}
	}

	targets := []*Target{}
	for ruleName, URLs := range c.Targets {
		rule, ok := c.Rules[ruleName]
		if !ok {
			return nil, fmt.Errorf("unknown rule %s", ruleName)
		}
		for _, rawURL := range URLs {
			parsedURL, err := url.Parse(rawURL)
			if err != nil {
				return nil, err
			}
			targets = append(targets, &Target{
				Name: rawURL,
				URL:  parsedURL,
				Rule: rule,
			})
		}
	}
	return targets, nil
}

func (r *Rule) Setup(metricMaker MetricMaker) error {
	if r.MetricName == "" {
		r.MetricName = DefaultMetricName
	}
	if r.Timeout == 0 {
		r.Timeout = DefaultTimeout
	}
	metricMaker.MakeMetric(r.MetricName)

	switch r.Type {
	case "http":
		if r.HTTPRule == nil {
			r.HTTPRule = &HTTPRule{}
		}
		err := r.HTTPRule.Setup()
		return err
	case "tcp":
		return nil
	default:
		return fmt.Errorf("unsupported type %s, expected http or tcp", r.Type)
	}
}

func (r *HTTPRule) Setup() error {
	if r.BodyRegex != "" {
		if r.BodyContent != "" {
			return fmt.Errorf("body_regexp and body_content are mutually exclusive")
		}
		var err error
		r.CompiledRegex, err = regexp.Compile(r.BodyRegex)
		if err != nil {
			return fmt.Errorf("cannot compile regex %s, %v", r.BodyRegex, err)
		}
	}
	if r.ValidHTTPStatuses != nil && len(r.ValidHTTPStatuses) > 0 && r.IgnoreHTTPStatus {
		return fmt.Errorf("ignore_http_status and statuses are mutually exclusive")
	}
	if r.BodyContent != "" {
		r.BodyContentBytes = []byte(r.BodyContent)
	}
	if r.ReadMax == 0 {
		r.ReadMax = DefaultReadMax
	}
	return nil
}
