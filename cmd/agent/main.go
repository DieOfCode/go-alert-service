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
	poll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	report := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
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

loop:
	for {
		select {
		case <-poll.C:
			a.CollectMetrics()
			logger.Info().Interface("metrics", a.Metrics).Msg("Metrics collected")
		case <-report.C:
			a.Retry(ctx, 3, func(ctx context.Context) error {
				return a.SendAllMetrics(ctx)
			}, 1*time.Second, 3*time.Second, 5*time.Second)

			logger.Info().Interface("metrics", a.Metrics).Msg("Metrics sent")
		case <-ctx.Done():
			logger.Info().Err(ctx.Err()).Send()
			poll.Stop()
			report.Stop()
			break loop
		}
	}
	logger.Info().Msg("Finished collecting metrics")
}
