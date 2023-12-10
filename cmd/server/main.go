package main

import (
	"fmt"
	"net/http"

	"github.com/DieOfCode/go-alert-service/internal/handler"
	"github.com/DieOfCode/go-alert-service/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	memStorage := storage.NewMemStorage()
	handler := handler.NewHandler(memStorage)

	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.MethodFunc(http.MethodPost, "/update/{type}/{name}/{value}", handler.HandleUpdateMetric)
		r.MethodFunc(http.MethodGet, "/update/{type}/{name}", handler.HandleGetMetricByName)
		r.MethodFunc(http.MethodGet, "/", handler.HandleGetAllMetrics)
	})

	err := http.ListenAndServe(":8080", router)

	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
