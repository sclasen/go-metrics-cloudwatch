package config

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"log"
	"time"
)

const (
	Perc50  = float64(0.50)
	Perc75  = float64(0.50)
	Perc95  = float64(0.50)
	Perc99  = float64(0.99)
	Perc999 = float64(0.999)
	Perc100 = float64(1)
)

type PutMetricsClient interface {
	PutMetricData(*cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error)
}

type Config interface {
	Namespace() string
	ShouldReport(metricName string) bool
	ReportingInterval() time.Duration
	Percentiles() []float64
	Client() PutMetricsClient
}

type StaticConfig struct {
	Cloudwatch  PutMetricsClient
	Interval    time.Duration
	Percs       []float64
	CWNamespace string
	Filter      func(string) bool
}

func (s *StaticConfig) ShouldReport(metricName string) bool {
	should := s.Filter(metricName)
	log.Printf("at=should-report metric=%s report=%t", metricName, should)
	return should
}

func (s *StaticConfig) ReportingInterval() time.Duration {
	return s.Interval
}

func (s *StaticConfig) Client() PutMetricsClient {
	return s.Cloudwatch
}

func (s *StaticConfig) Percentiles() []float64 {
	return s.Percs
}

func (s *StaticConfig) Namespace() string {
	return s.CWNamespace
}

/*
type DynamoDBConfig struct {
	globalEnabledMetrics []string
	perInstanceEnabledMetrics map[string]string
}

func (d *DynamodbConfig) PollConfig() {
	poll once every few minutes, read enabled metrics
}
*/
