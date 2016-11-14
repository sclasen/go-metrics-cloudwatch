package reporter

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/rcrowley/go-metrics"
	"github.com/sclasen/go-metrics-cloudwatch/config"
)

//blocks, run as go reporter.Cloudwatch(cfg)
func Cloudwatch(registry metrics.Registry, cfg *config.Config) {
	ticks := time.NewTicker(cfg.ReportingInterval)
	defer ticks.Stop()
	for {
		select {
		case <-ticks.C:
			emitMetrics(registry, cfg)
		}
	}
}

func emitMetrics(registry metrics.Registry, cfg *config.Config) {
	data := metricsData(registry, cfg)

	//20 is the max metrics per request
	for len(data) > 20 {
		put := data[0:20]
		putMetrics(cfg, put)
		data = data[20:]
	}

	if len(data) > 0 {
		putMetrics(cfg, data)
	}

}

func putMetrics(cfg *config.Config, data []*cloudwatch.MetricDatum) {
	client := cfg.Client
	req := &cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(cfg.Namespace),
		MetricData: data,
	}
	_, err := client.PutMetricData(req)
	if err != nil {
		log.Printf("component=cloudwatch-reporter fn=emitMetrics at=error error=%s", err)
	} else {
		log.Printf("component=cloudwatch-reporter fn=emitMetrics at=put-metrics count=%d", len(req.MetricData))
	}
}

func metricsData(registry metrics.Registry, cfg *config.Config) []*cloudwatch.MetricDatum {
	counters, gagues, histos, meters, timers := 0, 0, 0, 0, 0
	countersOut, gaguesOut, histosOut, metersOut, timersOut := 0, 0, 0, 0, 0

	data := []*cloudwatch.MetricDatum{}
	timestamp := aws.Time(time.Now())

	aDatum := func(name string) *cloudwatch.MetricDatum {
		return &cloudwatch.MetricDatum{
			MetricName: aws.String(name),
			Timestamp:  timestamp,
			Dimensions: dimensions(cfg),
		}
	}
	//rough port from the graphite reporter
	registry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			counters += 1
			count := float64(metric.Count())
			if cfg.Filter.ShouldReport(name, count) {
				datum := aDatum(name)
				datum.Unit = aws.String(cloudwatch.StandardUnitCount)
				datum.Value = aws.Float64(count)
				data = append(data, datum)
				countersOut +=1
			}
			if cfg.ResetCountersOnReport {
				metric.Clear()
			}
		case metrics.Gauge:
			gagues += 1
			value := float64(metric.Value())
			if cfg.Filter.ShouldReport(name, value) {
				datum := aDatum(name)
				datum.Unit = aws.String(cloudwatch.StandardUnitCount)
				datum.Value = aws.Float64(float64(value))
				data = append(data, datum)
				gaguesOut += 1
			}
		case metrics.GaugeFloat64:
			gagues += 1
			value := float64(metric.Value())
			if cfg.Filter.ShouldReport(name, value) {
				datum := aDatum(name)
				datum.Unit = aws.String(cloudwatch.StandardUnitCount)
				datum.Value = aws.Float64(value)
				data = append(data, datum)
				gaguesOut += 1
			}
		case metrics.Histogram:
			histos += 1
			h := metric.Snapshot()
			value := float64(h.Count())
			if cfg.Filter.ShouldReport(name, value) {
				for _, p := range cfg.Filter.Percentiles(name) {
					log.Printf("%+v", h)
					pname := fmt.Sprintf("%s-perc%.3f", name, p)
					pvalue := h.Percentile(p)
					if cfg.Filter.ShouldReport(pname, pvalue) {
						datum := aDatum(pname)
						datum.Unit = aws.String(cloudwatch.StandardUnitCount)
						datum.Value = aws.Float64(pvalue)
						data = append(data, datum)
						histosOut +=1
					}
				}
			}
		case metrics.Meter:
			meters += 1
			m := metric.Snapshot()
			dataz := map[string]float64{
				fmt.Sprintf("%s.count", name):          float64(m.Count()),
				fmt.Sprintf("%s.one-minute", name):     m.Rate1(),
				fmt.Sprintf("%s.five-minute", name):    m.Rate5(),
				fmt.Sprintf("%s.fifteen-minute", name): m.Rate15(),
				fmt.Sprintf("%s.mean", name):           m.RateMean(),
			}
			for n, v := range dataz {
				if cfg.Filter.ShouldReport(n, v) {
					datum := aDatum(n)
					datum.Value = aws.Float64(v)
					data = append(data, datum)
					metersOut +=1
				}
			}
		case metrics.Timer:
			timers += 1
			t := metric.Snapshot()
			if t.Count() == 0 {
				return
			}
			dataz := map[string]float64{
				fmt.Sprintf("%s.count", name):          float64(t.Count()),
				fmt.Sprintf("%s.one-minute", name):     t.Rate1(),
				fmt.Sprintf("%s.five-minute", name):    t.Rate5(),
				fmt.Sprintf("%s.fifteen-minute", name): t.Rate15(),
				fmt.Sprintf("%s.mean", name):           t.RateMean(),
			}
			for n, v := range dataz {
				if cfg.Filter.ShouldReport(n, v) {
					datum := aDatum(n)
					datum.Value = aws.Float64(v)
					data = append(data, datum)
					timersOut += 1
				}
			}
			for _, p := range cfg.Filter.Percentiles(name) {
				pname := fmt.Sprintf("%s-perc%.3f", name, p)
				pvalue := t.Percentile(p)
				if cfg.Filter.ShouldReport(pname, pvalue) {
					datum := aDatum(pname)
					datum.Value = aws.Float64(pvalue)
					data = append(data, datum)
					timersOut +=1
				}
			}

		}
	})
        total := counters + gagues + histos + meters + timers
	totalOut := countersOut + gaguesOut + histosOut + metersOut + timersOut
	log.Printf("component=cloudwatch-reporter fn=metricsData at=sources total=%d counters=%d gagues=%d histos=%d meters=%d timers=%d", total, counters, gagues, histos, meters, timers)
	log.Printf("component=cloudwatch-reporter fn=metricsData at=targets total=%d counters=%d gagues=%d histos=%d meters=%d timers=%d", totalOut, countersOut, gaguesOut, histosOut, metersOut, timersOut)

	return data
}

func dimensions(c *config.Config) []*cloudwatch.Dimension {
	ds := []*cloudwatch.Dimension{}
	for k, v := range c.StaticDimensions {
		d := &cloudwatch.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		}

		ds = append(ds, d)
	}
	return ds
}
