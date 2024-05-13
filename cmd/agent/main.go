package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DieOfCode/go-alert-service/internal/agent"
	"github.com/DieOfCode/go-alert-service/internal/configuration"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := configuration.NewAgent()
	if err != nil {
		logger.Fatal().Err(err).Msg("Configuration error")
	}

	client := &http.Client{
		Timeout: time.Minute,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	a := agent.New(&logger, client, cfg)

	logger.Info().
		Int("pollInterval", cfg.PollInterval).
		Int("reportInterval", cfg.ReportInterval).
		Msg("Started collecting metrics")

	go a.CollectRuntimeMetrics(ctx, time.Duration(cfg.PollInterval))
	go a.CollectGopsutilMetrics(ctx, time.Duration(cfg.PollInterval))
	metricsChan := a.PrepareMetrics(ctx, time.Duration(cfg.ReportInterval))
	for i := 0; i < cfg.RateLimit; i++ {
		go a.Retry(ctx, 3, func(ct context.Context) error {
			return a.SendMetrics(ct, metricsChan)
		})
	}
	logger.Info().Msg("Finished collecting metrics")
}
