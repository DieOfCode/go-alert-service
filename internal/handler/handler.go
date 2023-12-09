package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
	s "github.com/DieOfCode/go-alert-service/internal/storage"
)

type Handler struct {
	repository s.Repository
}

func NewHandler(repository s.Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (m *Handler) HandleUpdateMetric(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

	if len(parts) != 3 {
		http.Error(w, "Попытка передать запрос без имени метрики", http.StatusNotFound)
		return
	}

	metricType := metrics.MetricType(parts[0])
	metricName := parts[1]
	metricValue := parts[2]

	err := m.repository.UpdateMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Метрика успешно обновлена")

}
