go-metrics-cloudwatch
------------------

Reports go-metrics to cloudwatch.

usage
=====

```go

import (
    "github.com/sclasen/go-metrics-cloudwatch/config"
    "github.com/sclasen/go-metrics-cloudwatch/reporter"
)

metricsConf := &config.Config{
		Client:            cloudwatch.New(...),
		Namespace:         "my-metrics-namespace",
		Filter:            &config.NoFilter{},
		ReportingInterval: 1 * time.Minute,
		StaticDimensions:  []map[string]string{"name":"value"},
	}
go reporter.Cloudwatch(Registry, metricsConf)

```