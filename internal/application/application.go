package application

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DieOfCode/go-alert-service/internal/configuration"
	"github.com/DieOfCode/go-alert-service/internal/handler"
	"github.com/DieOfCode/go-alert-service/internal/repository"
	s "github.com/DieOfCode/go-alert-service/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func Run() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := configuration.NewServer()
	if err != nil {
		logger.Fatal().Err(err).Msg("Configuration error")
		return
	}

	var storage repository.Storage
	var db *sql.DB
	logger.Info().Msg(cfg.DatabaseDSN)
	if cfg.DatabaseDSN != "" {
		db, err = connectDB(&logger, &cfg)
		if err != nil {
			logger.Error().Err(err).Msg("DB initializing error")
			return
		}
		defer db.Close()
		storage = s.NewDatabaseStorage(&logger, db)
	} else {
		storage = s.NewMemStorage(&logger, *cfg.StoreInterval, cfg.FileStoragePath)

	}

	repository := repository.New(&logger, storage)

	server := NewServer(&logger, cfg.ServerAddress, repository, db)
	server.RegisterHandler()
	if *cfg.Restore {
		err := storage.RestoreFromFile()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to restore storage from file")
		}
		logger.Info().Msg("Storage has been restored from file")
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

	go server.ListenAndServe(&cfg)

	<-ctx.Done()
	logger.Info().Msg("Shutdown signal received")

	if err := storage.WriteToFile(); err != nil {
		logger.Error().Err(err).Msg("Failed to write storage content to file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.server.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Shutdown server error")
	}

	logger.Info().Msg("Server stopped gracefully")
}

func connectDB(logger *zerolog.Logger, cfg *configuration.Config) (*sql.DB, error) {
	print("inside bd init")
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		logger.Fatal().Err(err).Msg("DB initializing error")
		return nil, err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		print("proble getting instance")
		logger.Fatal().Err(err).Msg("DB pinging error")
		return nil, err
	}

	if _, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS metrics (
		id VARCHAR NOT NULL,
		type VARCHAR NOT NULL,
		delta BIGINT,
		value DOUBLE PRECISION,
		UNIQUE (id, type)
	)`); err != nil {
		log.Fatalf("Error creating metrics table: %v", err)
		return nil, err
	}
	return db, nil
}

func DBPing(logger *zerolog.Logger, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info().Msg("Start DB PIMG")
		if db == nil {
			logger.Info().Msg("Dont have DB")
			return
		}
		if err := db.Ping(); err != nil {
			logger.Error().Err(err).Msg("Pinging DB error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

type Server struct {
	server *http.Server
	logger *zerolog.Logger
	repo   *repository.Repository
	db     *sql.DB
}

func NewServer(l *zerolog.Logger, addr string, repo *repository.Repository, db *sql.DB) *Server {
	return &Server{
		server: &http.Server{Addr: addr},
		logger: l,
		repo:   repo,
		db:     db,
	}
}

func (server *Server) RegisterHandler() {
	metricHandler := handler.NewMetricHandler(server.logger, server.repo)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&handler.LogFormatter{Logger: server.logger}))
		r.Use(middleware.Compress(5, "text/html", "application/json"))
		r.Use(handler.Decompress(server.logger))
		r.Use(middleware.Recoverer)
		r.MethodFunc(http.MethodPost, "/update/{type}/{name}/{value}", metricHandler.SaveMetric)
		r.MethodFunc(http.MethodGet, "/value/{type}/{name}", metricHandler.GetMetricByName)
		r.MethodFunc(http.MethodGet, "/", metricHandler.GetAllMetrics)
		r.MethodFunc(http.MethodPost, "/update/", metricHandler.SaveMetricWithJSON)
		r.MethodFunc(http.MethodPost, "/updates/", metricHandler.SaveMetricsWithJSON)
		r.MethodFunc(http.MethodPost, "/value/", metricHandler.GetMetricByNameWithJSON)
		r.Method(http.MethodGet, "/ping", DBPing(server.logger, server.db))
	})
	server.server.Handler = r
}

func (server *Server) ListenAndServe(cfg *configuration.Config) {
	server.logger.Info().Msgf("Server is listerning on %s", cfg.ServerAddress)
	if err := server.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		server.logger.Error().Err(err).Msg("Server error")
		return
	}
}
