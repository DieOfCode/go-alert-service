package storage

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
)

type MemStorage struct {
	mu      sync.Mutex
	metrics map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]metrics.Metric),
	}
}

func (m *MemStorage) UpdateMetric(metricType metrics.MetricType, metricName string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s_%s", metricType, metricName)

	switch metricType {
	case metrics.Gauge:
		if newValue, err := strconv.ParseFloat(value, 64); err == nil {
			m.metrics[key] = metrics.Metric{Value: newValue}
		} else {
			return fmt.Errorf("некорректное значение для типа counter: %v", value)

		}

	case metrics.Counter:
		if existingMetric, ok := m.metrics[key]; ok {
			switch existingValue := existingMetric.Value.(type) {
			case int64:
				if newValue, err := strconv.ParseInt(value, 10, 64); err == nil {
					m.metrics[key] = metrics.Metric{Value: existingValue + newValue}
				} else {
					return fmt.Errorf("некорректное значение для типа counter: %v", value)
				}
			default:
				return fmt.Errorf("некорректное предыдущее значение для типа counter: %v", existingMetric.Value)
			}
		} else {
			m.metrics[key] = metrics.Metric{Value: value}
		}
	default:
		return fmt.Errorf("некорректный тип метрики: %s", metricType)
	}

	return nil
}

func (m *MemStorage) HandleUpdateMetric(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

	if len(parts) != 3 {
		http.Error(w, "Попытка передать запрос без имени метрики", http.StatusNotFound)
		return
	}

	metricType := metrics.MetricType(parts[0])
	metricName := parts[1]
	metricValue := parts[2]

	err := m.UpdateMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Метрика успешно обновлена")
}
