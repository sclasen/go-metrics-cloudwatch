package reporter

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
	"github.com/sclasen/go-metrics-cloudwatch/config"
	"testing"
)

type MockPutMetricsClient struct {
	metricsPut int
}

func (m *MockPutMetricsClient) PutMetricData(in *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	m.metricsPut += len(in.MetricData)
	return &cloudwatch.PutMetricDataOutput{}, nil
}

func TestCloudwatchReporter(t *testing.T) {
	mock := &MockPutMetricsClient{}
	cfg := &config.Config{
		Client: mock,
		Filter: &config.NoFilter{},
	}
	registry := metrics.DefaultRegistry
	count := metrics.GetOrRegisterCounter("count", registry)
	count.Inc(1)

	emitMetrics(registry, cfg)

	if mock.metricsPut < 1 {
		t.Fatal("No Metrics Put")
	}
}
