package main

import (
	"net/http"
	"os"

	"github.com/DieOfCode/go-alert-service/internal/configuration"
	"github.com/DieOfCode/go-alert-service/internal/handler"
	log "github.com/DieOfCode/go-alert-service/internal/logger"
	"github.com/DieOfCode/go-alert-service/internal/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	config := configuration.ServerConfiguration()
	memStorage := storage.NewMemStorage(logger)
	handler := handler.NewHandler(memStorage, logger)

	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Use(middleware.RequestLogger(&log.LogFormatter{Logger: &logger}))
		r.Use(middleware.Recoverer)
		r.Use(middleware.Compress(5, "text/html",
			"application/json"))
		r.Use(handler.Decompress())
		r.Use(middleware.Recoverer)
		r.MethodFunc(http.MethodPost, "/update/{type}/{name}/{value}", handler.HandleUpdateMetric)
		r.MethodFunc(http.MethodGet, "/value/{type}/{name}", handler.HandleGetMetricByName)
		r.MethodFunc(http.MethodGet, "/", handler.HandleGetAllMetrics)
		r.MethodFunc(http.MethodPost, "/update/", handler.HandleUpdateJSONMetric)
		r.MethodFunc(http.MethodPost, "/value/", handler.HandleGetJSONMetric)
	})

	err := http.ListenAndServe(config.ServerAddress, router)
	logger.Info().Msgf("Server is listerning on %s", config.ServerAddress)
	if err != nil {
		logger.Fatal().Err(err).Msg("Server start error")
	}
}
