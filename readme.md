go-metrics-cloudwatch
------------------

Reports go-metrics to cloudwatch.

[![Build Status](https://api.travis-ci.org/sclasen/go-metrics-cloudwatch.svg?branch=master)](https://travis-ci.org/sclasen/go-metrics-cloudwatch)

usage
=====

```go

import (
    "github.com/sclasen/go-metrics-cloudwatch/config"
    "github.com/sclasen/go-metrics-cloudwatch/reporter"
    "github.com/aws/aws-sdk-go/service/cloudwatch"
    "github.com/rcrowley/go-metrics"
)

registry := metrics.NewRegistry()
metricsConf := &config.Config{
		Client:            cloudwatch.New(...),
		Namespace:         "my-metrics-namespace",
		Filter:            &config.NoFilter{},
		ReportingInterval: 1 * time.Minute,
		StaticDimensions:  []map[string]string{"name":"value"},
	}
go reporter.Cloudwatch(registry, metricsConf)

```