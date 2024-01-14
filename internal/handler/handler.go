package handler

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

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

func (handler *Handler) HandleUpdateMetric(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Попытка передать запрос без имени метрики", http.StatusNotFound)
		return
	}

	metricType := metrics.MetricType(parts[0])
	metricName := parts[1]
	metricValue := parts[2]

	err := handler.repository.UpdateMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Метрика успешно обновлена")

}

func (handler *Handler) HandleGetMetricByName(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/value/"), "/")

	if len(parts) != 2 {
		http.Error(w, "Попытка передать запрос без имени метрики", http.StatusNotFound)
		return
	}
	metricType := metrics.MetricType(parts[0])
	metricName := parts[1]

	metric, err := handler.repository.GetMetricByName(metricType, metricName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	print("МЕТРИКА ПОЛУЧЕНА")
	w.Header().Set("Content-Type", "text/plain")
	if metricType == metrics.Gauge {
		w.Write([]byte(strconv.FormatFloat(metric.Value.(float64), 'f', -1, 64)))
	}
	if metricType == metrics.Counter {
		w.Write([]byte(fmt.Sprintf("%d", metric.Value.(int64))))
	}
	w.WriteHeader(http.StatusOK)

}

func (handler *Handler) HandleGetAllMetrics(w http.ResponseWriter, r *http.Request) {
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

	err = template.Execute(w, handler.repository.GetAllMetrics())

	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return
	}

}

func (handler *Handler) HandleUpdateJSONMetric(w http.ResponseWriter, r *http.Request) {
	var metric metrics.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		handler.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, error.Error{Error: "Bad request"})
		return
	}
	handler.logger.Info().Any("req", metric).Msg("UPDATE Decoded request body")
	var metricValue string
	var metricType metrics.MetricType
	if metric.MType == "gauge" {
		metricType = metrics.Gauge
		metricValue = strconv.FormatFloat(float64(*metric.Value), 'f', 10, 64)
	} else {
		metricType = metrics.Counter
		metricValue = fmt.Sprintf("%d", *metric.Delta)
	}

	if err := handler.repository.UpdateMetric(metricType, metric.ID, metricValue); err != nil {
		handler.logger.Error().Err(err).Msg("UpdateMetric method error")
		writeResponse(w, http.StatusInternalServerError, error.Error{Error: "Internal server error"})
		return
	}

	writeResponse(w, http.StatusOK, metric)
	handler.logger.Info().Any("req", r.Body).Any("MetricName", metric.ID).Any("MetricType", metric.MType).Any("MetricValue", metricValue).Any("GaudeValue", metric.Value).Msg("Success save metric")
}

func (handler *Handler) HandleGetJSONMetric(w http.ResponseWriter, r *http.Request) {

	handler.logger.Info().Any("req", r.Body).Msg("Request body")

	var metric metrics.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		handler.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, error.Error{Error: "Bad request"})
		return
	}
	handler.logger.Info().Any("req", metric).Msg("Decoded request body")

	var metricType metrics.MetricType
	if metric.MType == "gauge" {
		metricType = metrics.Gauge
	} else {
		metricType = metrics.Counter
	}
	res, err := handler.repository.GetMetricByName(metricType, metric.ID)
	if err != nil {
		handler.logger.Error().Err(err).Msg("GetMetricByName method error")
		writeResponse(w, http.StatusNotFound, error.Error{Error: "Not found"})
		return
	}
	var resultMetric metrics.Metrics

	if metricType == metrics.Gauge {
		value := res.Value.(float64)
		resultMetric = metrics.Metrics{ID: metric.ID, MType: metric.MType, Value: &value}
	} else {
		value := res.Value.(int64)
		resultMetric = metrics.Metrics{ID: metric.ID, MType: metric.MType, Delta: &value}
	}

	writeResponse(w, http.StatusOK, resultMetric)
	handler.logger.Info().Any("req", r.Body).Any("MetricName", res.MetricName).Any("MetricValue", res.Value).Msg("Success get metric")
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

func (handler *Handler) Decompress() func(next http.Handler) http.Handler {
	pool := sync.Pool{
		New: func() any { return new(gzip.Reader) },
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encodingHeaders := r.Header.Values("Content-Encoding")
			if !slices.Contains(encodingHeaders, "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gr, ok := pool.Get().(*gzip.Reader)
			if !ok {
				handler.logger.Error().Msg("Error to get Reader")
			}
			defer pool.Put(gr)

			if err := gr.Reset(r.Body); err != nil {
				handler.logger.Error().Err(err).Msg("Reset gr error")
			}
			defer gr.Close()

			r.Body = gr
			next.ServeHTTP(w, r)
		})
	}
}
