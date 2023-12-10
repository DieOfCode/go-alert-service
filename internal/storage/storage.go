package storage

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
)

type MemStorage struct {
	mu      sync.Mutex
	metrics map[string]metrics.Metric
}

type Repository interface {
	UpdateMetric(metricType metrics.MetricType, metricName string, value string) error
	GetMetricByName(metricType metrics.MetricType, metricName string) (metrics.Metric, error)
	GetAllMetrics() []metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]metrics.Metric),
	}
}

func (storage *MemStorage) UpdateMetric(metricType metrics.MetricType, metricName string, value string) error {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	key := fmt.Sprintf("%s_%s", metricType, metricName)
	fmt.Print("KEY\n")
	fmt.Print(key)
	switch metricType {
	case metrics.Gauge:
		if newValue, err := strconv.ParseFloat(value, 64); err == nil {
			storage.metrics[key] = metrics.Metric{Value: newValue}
		} else {
			return fmt.Errorf("некорректное значение для типа counter: %v", value)

		}

	case metrics.Counter:
		if existingMetric, ok := storage.metrics[key]; ok {
			switch existingValue := existingMetric.Value.(type) {
			case int64:
				if newValue, err := strconv.ParseInt(value, 10, 64); err == nil {
					storage.metrics[key] = metrics.Metric{Value: existingValue + newValue}
				} else {
					return fmt.Errorf("некорректное значение для типа counter: %v", value)
				}
			default:
				return fmt.Errorf("некорректное предыдущее значение для типа counter: %v", existingMetric.Value)
			}
		} else {
			if newValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				storage.metrics[key] = metrics.Metric{Value: newValue}
			} else {
				return fmt.Errorf("некорректное значение для типа counter: %v", value)
			}

		}
	default:
		return fmt.Errorf("некорректный тип метрики: %s", metricType)
	}

	return nil
}

func (storage *MemStorage) GetMetricByName(metricType metrics.MetricType, metricName string) (metrics.Metric, error) {
	key := fmt.Sprintf("%s_%s", metricType, metricName)
	metric, ok := storage.metrics[key]
	if !ok {
		return metrics.Metric{}, fmt.Errorf("метрика с именем %s не найдена", key)
	}
	return metric, nil
}

func (storage *MemStorage) GetAllMetrics() []metrics.Metric {
	return getAllValues(storage.metrics)
}

func getAllValues(metricsByName map[string]metrics.Metric) []metrics.Metric {
	values := make([]metrics.Metric, len(metricsByName))

	for _, value := range metricsByName {
		values = append(values, value)
	}
	return values
}
