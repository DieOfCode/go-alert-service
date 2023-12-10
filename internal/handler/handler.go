package handler

import (
	"fmt"
	"html/template"
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
	metrics := m.repository.GetAllMetrics()
	fmt.Println("METRICS")
	fmt.Println(metrics)
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
	metricType := metrics.MetricType(parts[0])
	metricName := parts[1]

	metric, err := m.repository.GetMetricByName(metricType, metricName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if metricType == metrics.Gauge {
		w.Write([]byte(fmt.Sprintf("%.3f", metric.Value.(float64))))
	}
	if metricType == metrics.Gauge {
		w.Write([]byte(fmt.Sprintf("%d", metric.Value.(int64))))
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
