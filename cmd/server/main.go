package main

import (
	"fmt"
	"net/http"

	"github.com/DieOfCode/go-alert-service/internal/configuration"
	"github.com/DieOfCode/go-alert-service/internal/handler"
	"github.com/DieOfCode/go-alert-service/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {

	config := configuration.ServerConfiguration()
	memStorage := storage.NewMemStorage()
	handler := handler.NewHandler(memStorage)

	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.MethodFunc(http.MethodPost, "/update/{type}/{name}/{value}", handler.HandleUpdateMetric)
		r.MethodFunc(http.MethodGet, "/value/{type}/{name}", handler.HandleGetMetricByName)
		r.MethodFunc(http.MethodGet, "/", handler.HandleGetAllMetrics)
	})

	err := http.ListenAndServe(config.ServerAddress, router)

	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
