package storage

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
)

type MemStorage struct {
	mu      sync.Mutex
	metrics map[string]metrics.Metrics
}

type Repository interface {
	UpdateMetric(metricType string, metricName string, value string) error
	GetMetricByName(metricType string, metricName string) (metrics.Metrics, error)
	GetAllMetrics() []metrics.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]metrics.Metrics),
	}
}

func (storage *MemStorage) UpdateMetric(metricType string, metricName string, value string) error {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	key := metricName
	fmt.Print("KEY\n")
	fmt.Print(key)
	switch metricType {
	case metrics.Gauge:
		if newValue, err := strconv.ParseFloat(value, 64); err == nil {
			storage.metrics[key] = metrics.Metrics{Value: &newValue}
		} else {
			return fmt.Errorf("некорректное значение для типа counter: %v", value)

		}

	case metrics.Counter:
		if existingMetric, ok := storage.metrics[key]; ok {
			existingValue := existingMetric.Delta
			newValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("некорректное значение для типа counter: %v", value)

			}

			firstSum := *existingValue
			updatedValue := firstSum + newValue
			storage.metrics[key] = metrics.Metrics{Delta: &updatedValue}

		} else {
			if newValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				storage.metrics[key] = metrics.Metrics{Delta: &newValue}
			} else {
				return fmt.Errorf("некорректное значение для типа counter: %v", value)
			}

		}
	default:
		return fmt.Errorf("некорректный тип метрики: %s", metricType)
	}

	return nil
}

func (storage *MemStorage) GetMetricByName(metricType string, metricName string) (metrics.Metrics, error) {
	key := metricName
	metric, ok := storage.metrics[key]
	if !ok {
		return metrics.Metrics{}, fmt.Errorf("метрика с именем %s не найдена", key)
	}
	return metric, nil
}

func (storage *MemStorage) GetAllMetrics() []metrics.Metrics {
	return getAllValues(storage.metrics)
}

func getAllValues(metricsByName map[string]metrics.Metrics) []metrics.Metrics {
	values := make([]metrics.Metrics, len(metricsByName))

	for _, value := range metricsByName {
		values = append(values, value)
	}
	return values
}
