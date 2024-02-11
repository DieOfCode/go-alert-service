package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/DieOfCode/go-alert-service/internal/configuration"
	"github.com/DieOfCode/go-alert-service/internal/handler"
	"github.com/DieOfCode/go-alert-service/internal/repository"
	"github.com/DieOfCode/go-alert-service/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := configuration.NewServer()
	if err != nil {
		logger.Fatal().Err(err).Msg("Configuration error")
	}

	var db *sql.DB
	if cfg.DatabaseDNS != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDNS)
		if err != nil {
			logger.Fatal().Err(err).Msg("DB initializing error")
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			logger.Fatal().Err(err).Msg("DB pinging error")
		}
	}

	storage := storage.New(&logger, *cfg.StoreInterval, cfg.FileStoragePath)
	repository := repository.New(&logger, storage)
	metricHandler := handler.NewMetricHandler(&logger, repository)

	if *cfg.Restore {
		err := storage.RestoreFromFile()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to restore storage from file")
		}
		logger.Info().Msg("Storage has been restored from file")
	}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&handler.LogFormatter{Logger: &logger}))
		r.Use(middleware.Compress(5, "text/html", "application/json"))
		r.Use(handler.Decompress(&logger))
		r.Use(middleware.Recoverer)
		r.MethodFunc(http.MethodPost, "/update/{type}/{name}/{value}", metricHandler.SaveMetric)
		r.MethodFunc(http.MethodGet, "/value/{type}/{name}", metricHandler.GetMetricByName)
		r.MethodFunc(http.MethodGet, "/", metricHandler.GetAllMetrics)
		r.MethodFunc(http.MethodPost, "/update/", metricHandler.SaveMetricWithJson)
		r.MethodFunc(http.MethodPost, "/value/", metricHandler.GetMetricByNameWithJson)
		r.Method(http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if db == nil {
				return
			}
			if err := db.Ping(); err != nil {
				logger.Error().Err(err).Msg("Pinging DB error")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
	})

	server := http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if *cfg.StoreInterval > 0 {
		ticker := time.NewTicker(time.Duration(*cfg.StoreInterval) * time.Second)

		go func() {
		loop:
			for {
				select {
				case <-ticker.C:
					if err := storage.WriteToFile(); err != nil {
						logger.Error().Err(err).Msg("Failed to write storage content to file")
					}
				case <-ctx.Done():
					ticker.Stop()
					break loop
				}
			}
		}()
	}

	go func() {
		logger.Info().Msgf("Server is listerning on %s", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	<-ctx.Done()
	logger.Info().Msg("Shutdown signal received")

	if err := storage.WriteToFile(); err != nil {
		logger.Error().Err(err).Msg("Failed to write storage content to file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Shutdown server error")
	}

	logger.Info().Msg("Server stopped gracefully")
}
