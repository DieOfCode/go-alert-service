package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/DieOfCode/go-alert-service/internal/agent"
	m "github.com/DieOfCode/go-alert-service/internal/metrics"
)

func main() {
	parseFlags()
	httpClient := &http.Client{
		Timeout: time.Minute,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	poolTicker := time.NewTicker(time.Duration(poolInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	var metrics []m.Metric
	var counter int64

loop:
	for {
		select {
		case <-reportTicker.C:
			metrics = append(metrics, m.Metric{MetricType: m.Counter, MetricName: m.PoolCount, Value: counter})
			err := agent.SendMetric(ctx, *httpClient, metrics, addressHttp)
			if err != nil {
				log.Fatal(err)
			}
		case <-poolTicker.C:
			counter++
			metrics = agent.CollectGaudeMetrics()
			metrics = append(metrics, m.Metric{MetricType: m.Gauge, MetricName: m.RandomValue, Value: rand.Float64()})

		case <-ctx.Done():

			poolTicker.Stop()
			reportTicker.Stop()
			break loop
		}
	}

}
