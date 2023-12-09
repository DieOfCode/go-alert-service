package main

import (
	"context"
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
			metrics = append(metrics, m.Metric{MetricType: m.Counter, MetricName: "PollCount", Value: counter})
			err := agent.SendMetric(ctx, *httpClient, metrics)
			if err != nil {
				//TODO handle error
			}
		case <-poolTicker.C:

		default:
			break loop
		}
	}

}
