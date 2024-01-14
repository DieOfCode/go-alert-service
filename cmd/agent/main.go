package main

import (
	"compress/gzip"
	"context"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DieOfCode/go-alert-service/internal/agent"
	"github.com/DieOfCode/go-alert-service/internal/configuration"
	m "github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	var metrics []m.Metric
	var counter int64

	config := configuration.AgentConfiguration()
	httpClient := &http.Client{
		Timeout: time.Minute,
	}
	metricAgent := agent.NewMetricAgent(logger)
	poolTicker := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	pool := &sync.Pool{
		New: func() any { return gzip.NewWriter(io.Discard) },
	}
	defer cancel()

	logger.Info().
		Int("pollInterval", config.PollInterval).
		Int("reportInterval", config.ReportInterval).
		Msg("Started collecting metrics")

loop:
	for {
		select {
		case <-reportTicker.C:
			metrics = append(metrics, m.Metric{MetricType: m.Counter, MetricName: m.PoolCount, Value: counter})
			err := metricAgent.SendMetric(ctx, httpClient, metrics, config.ServerAddress, pool)
			if err != nil {
				logger.Fatal().Err(err).Msg("Send metrics error")
			} else {
				logger.Info().Interface("metrics", metrics).Msg("Metrics sent")
			}
		case <-poolTicker.C:
			counter++
			metrics = metricAgent.CollectGaudeMetrics()
			metrics = append(metrics, m.Metric{MetricType: m.Gauge, MetricName: m.RandomValue, Value: rand.Float64()})
			logger.Info().Interface("metrics", metrics).Msg("Metrics collected")

		case <-ctx.Done():

			poolTicker.Stop()
			reportTicker.Stop()
			break loop
		}
	}

}
