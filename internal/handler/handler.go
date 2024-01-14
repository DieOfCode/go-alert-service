package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/DieOfCode/go-alert-service/internal/error"
	"github.com/DieOfCode/go-alert-service/internal/metrics"
	s "github.com/DieOfCode/go-alert-service/internal/storage"
	"github.com/rs/zerolog"
)

type Handler struct {
	repository s.Repository
	logger     zerolog.Logger
}

func NewHandler(repository s.Repository, logger zerolog.Logger) *Handler {
	return &Handler{
		repository: repository,
		logger:     logger,
	}
}

func (m *Handler) HandleUpdateMetric(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Попытка передать запрос без имени метрики", http.StatusNotFound)
		return
	}

	metricType := parts[0]
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

func (m *Handler) HandleGetMetricByName(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/value/"), "/")

	if len(parts) != 2 {
		http.Error(w, "Попытка передать запрос без имени метрики", http.StatusNotFound)
		return
	}
	metricType := parts[0]
	metricName := parts[1]

	metric, err := m.repository.GetMetricByName(metricType, metricName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	print("МЕТРИКА ПОЛУЧЕНА")
	w.Header().Set("Content-Type", "text/plain")
	if metricType == metrics.Gauge {
		w.Write([]byte(strconv.FormatFloat(*metric.Value, 'f', -1, 64)))
	}
	if metricType == metrics.Counter {
		w.Write([]byte(fmt.Sprintf("%d", *metric.Delta)))
	}
	w.WriteHeader(http.StatusOK)

}

func (m *Handler) HandleGetAllMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Metrics list</title>
	</head>
	<body>
		<h1>Metrics list</h1>
		<ul>
		{{range $name, $metric := .Metrics}}
			<li>{{$name}}: {{$metric.Value}}</li>
		{{end}}
		</ul>
	</body>
	</html>`
	template, err := template.New("metrics").Parse(tmpl)
	if err != nil {
		http.Error(w, "Ошибка создания шаблона", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")

	err = template.Execute(w, m.repository.GetAllMetrics())

	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return
	}

}

func (m *Handler) HandleUpdateJSONMetric(w http.ResponseWriter, r *http.Request) {
	var metric metrics.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		m.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, error.Error{Error: "Bad request"})
		return
	}
	var metricValue string
	var metricType string
	if metric.MType == "gauge" {
		metricType = metrics.Gauge
		metricValue = strconv.FormatFloat(float64(*metric.Value), 'f', 10, 64)
	} else {
		metricType = metrics.Counter
		metricValue = fmt.Sprintf("%d", *metric.Delta)
	}

	if err := m.repository.UpdateMetric(metricType, metric.ID, metricValue); err != nil {
		m.logger.Error().Err(err).Msg("UpdateMetric method error")
		writeResponse(w, http.StatusInternalServerError, error.Error{Error: "Internal server error"})
		return
	}

	writeResponse(w, http.StatusOK, metric)
	m.logger.Info().Any("req", r.Body).Any("MetricName", metric.ID).Any("MetricType", metric.MType).Any("MetricValue", metricValue).Any("GaudeValue", metric.Value).Msg("Success save metric")
}

func (m *Handler) HandleGetJSONMetric(w http.ResponseWriter, r *http.Request) {

	m.logger.Info().Any("req", r.Body).Msg("Request body")

	var metric metrics.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		m.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, error.Error{Error: "Bad request"})
		return
	}
	m.logger.Info().Any("req", metric).Msg("Decoded request body")

	var metricType string
	if metric.MType == "gauge" {
		metricType = metrics.Gauge
	} else {
		metricType = metrics.Counter
	}
	res, err := m.repository.GetMetricByName(metricType, metric.ID)
	if err != nil {
		m.logger.Error().Err(err).Msg("GetMetricByName method error")
		writeResponse(w, http.StatusNotFound, error.Error{Error: "Not found"})
		return
	}
	var resultMetric metrics.Metrics

	if metricType == metrics.Gauge {
		value := res.Value
		resultMetric = metrics.Metrics{ID: metric.ID, MType: metric.MType, Value: value}
	} else {
		value := res.Delta
		resultMetric = metrics.Metrics{ID: metric.ID, MType: metric.MType, Delta: value}
	}

	writeResponse(w, http.StatusOK, resultMetric)
	m.logger.Info().Any("req", r.Body).Any("MetricName", res.ID).Any("MetricValue", res.Value).Msg("Success get metric")
}

func writeResponse(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
		return
	}
	w.WriteHeader(code)
	w.Write(b)
}
