package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/DieOfCode/go-alert-service/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type Service interface {
	SaveMetric(m metrics.Metric) error
	GetMetric(mtype, mname string) (*metrics.Metric, error)
	GetMetrics() (metrics.Data, error)
}

type GetMetric struct {
	logger  *zerolog.Logger
	service Service
}

func NewGetMetric(l *zerolog.Logger, srv Service) *GetMetric {
	return &GetMetric{
		logger:  l,
		service: srv,
	}
}

func (h *GetMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	mname := chi.URLParam(r, "name")

	metric, err := h.service.GetMetric(mtype, mname)
	if err != nil {
		writeResponse(w, http.StatusNotFound, metrics.Error{Error: "Not found"})
		return
	}
	h.logger.Info().Any("metric", metric).Msg("Received metric from storage")

	switch mtype {
	case metrics.TypeGauge:
		writeResponse(w, http.StatusOK, *metric.Value)
	case metrics.TypeCounter:
		writeResponse(w, http.StatusOK, *metric.Delta)
	}
}

type GetMetricV2 struct {
	logger  *zerolog.Logger
	service Service
}

func NewGetMetricV2(l *zerolog.Logger, s Service) *GetMetricV2 {
	return &GetMetricV2{
		logger:  l,
		service: s,
	}
}

func (h *GetMetricV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Any("req", r.Body).Msg("Request body")

	var req metrics.Metric
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, metrics.Error{Error: "Bad request"})
		return
	}
	h.logger.Info().Any("req", req).Msg("Decoded request body")

	res, err := h.service.GetMetric(req.MType, req.ID)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetMetric method error")
		writeResponse(w, http.StatusNotFound, metrics.Error{Error: "Not found"})
		return
	}

	writeResponse(w, http.StatusOK, res)
}

type GetMetrics struct {
	logger  *zerolog.Logger
	service Service
}

func NewGetMetrics(l *zerolog.Logger, srv Service) *GetMetrics {
	return &GetMetrics{
		logger:  l,
		service: srv,
	}
}

func (h *GetMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allMetrics, err := h.service.GetMetrics()
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, metrics.Error{Error: "Internal server error"})
		return
	}
	h.logger.Info().Any("metrics", allMetrics).Msg("Received metrics from storage")

	tmpl, err := template.New("metrics").Parse(HTMLTemplateString)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, metrics.Error{Error: "Internal server error"})
		return
	}
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, allMetrics); err != nil {
		writeResponse(w, http.StatusInternalServerError, metrics.Error{Error: "Internal server error"})
		return
	}

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

const HTMLTemplateString = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    <ul>
    {{range .}}{{range .}}
        <li>ID: {{.ID}}, Value: {{.Value}}, Delta: {{.Delta}}</li>
    {{end}}{{end}}
    </ul>
</body>
</html>
`

type PostMetric struct {
	logger  *zerolog.Logger
	service Service
}

func NewPostMetric(l *zerolog.Logger, srv Service) *PostMetric {
	return &PostMetric{
		logger:  l,
		service: srv,
	}
}

func (h *PostMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "type")
	mname := chi.URLParam(r, "name")
	mvalue := chi.URLParam(r, "value")

	if mtype != metrics.TypeCounter && mtype != metrics.TypeGauge {
		print(fmt.Sprintf("\n 1 ,%s", mtype))
		print(fmt.Sprintf("\n 1.1 ,%s", r.URL.Path))
		writeResponse(w, http.StatusBadRequest, metrics.Error{Error: "Bad request"})
		return
	}

	var m metrics.Metric

	switch mtype {
	case metrics.TypeCounter:
		delta, err := strconv.ParseInt(mvalue, 10, 0)
		if err != nil {
			print("\n\n 1")
			writeResponse(w, http.StatusBadRequest, metrics.Error{Error: "Bad request"})
			return
		}
		m = metrics.Metric{
			ID:    mname,
			MType: mtype,
			Delta: &delta,
		}
	case metrics.TypeGauge:
		value, err := strconv.ParseFloat(mvalue, 64)
		if err != nil {
			print("\n\n 1")
			writeResponse(w, http.StatusBadRequest, metrics.Error{Error: "Bad request"})
			return
		}
		m = metrics.Metric{
			ID:    mname,
			MType: mtype,
			Value: &value,
		}
	}

	if err := h.service.SaveMetric(m); err != nil {
		if errors.Is(err, repository.ErrParseMetric) {
			print("\n\n 1")
			writeResponse(w, http.StatusBadRequest, metrics.Error{Error: "Bad request"})
			return
		}
		writeResponse(w, http.StatusInternalServerError, metrics.Error{Error: "Internal server error"})
		return
	}

	writeResponse(w, http.StatusOK, fmt.Sprintf("metric %s of type %s with value %v has been set successfully", mname, mtype, mvalue))
}

type PostMetricV2 struct {
	logger  *zerolog.Logger
	service Service
}

func NewPostMetricV2(l *zerolog.Logger, srv Service) *PostMetricV2 {
	return &PostMetricV2{
		logger:  l,
		service: srv,
	}
}

func (h *PostMetricV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Any("req", r.Body).Msg("Request body")

	var req metrics.Metric
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid incoming data")
		writeResponse(w, http.StatusBadRequest, metrics.Error{Error: "Bad request"})
		return
	}
	h.logger.Info().Any("req", req).Msg("Decoded request body")

	if err := h.service.SaveMetric(req); err != nil {
		h.logger.Error().Err(err).Msg("SaveMetric method error")
		writeResponse(w, http.StatusInternalServerError, metrics.Error{Error: "Internal server error"})
		return
	}

	writeResponse(w, http.StatusOK, req)
}

func writeResponse(w http.ResponseWriter, code int, v any) {
	w.Header().Add("Content-Type", "application/json")
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
		return
	}
	w.WriteHeader(code)
	w.Write(b)
}
