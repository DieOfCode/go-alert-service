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

var (
	reportInterval = time.Second * 10
	poolInterval   = time.Second * 5
)

func main() {

	httpClient := &http.Client{}

	poolTicker := time.NewTicker(poolInterval)
	reportTicker := time.NewTicker(reportInterval)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)

	var metrics []m.Metric
	var counter int64

	defer cancel()

loop:
	for {
		select {
		case <-reportTicker.C:
			metrics = append(metrics, m.Metric{MetricType: m.Counter, MetricName: m.PoolCount, Value: counter})
			err := agent.SendMetric(ctx, *httpClient, metrics)
			if err != nil {
				log.Fatal(err)
			}
		case <-poolTicker.C:
			counter++
			metrics = agent.CollectGaudeMetrics()
			metrics = append(metrics, m.Metric{MetricType: m.Gauge, MetricName: m.RandomValue, Value: rand.Float64()})

		default:

			poolTicker.Stop()
			reportTicker.Stop()
			break loop
		}
	}

}
