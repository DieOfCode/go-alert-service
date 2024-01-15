package storage

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/rs/zerolog"
)

type MemStorage struct {
	mu      sync.Mutex
	metrics map[string]metrics.Metrics
	logger  zerolog.Logger
}

type Repository interface {
	UpdateMetric(metricType string, metricName string, value string) error
	GetMetricByName(metricType string, metricName string) (metrics.Metrics, error)
	GetAllMetrics() []metrics.Metrics
}

func NewMemStorage(logger zerolog.Logger) *MemStorage {
	return &MemStorage{
		metrics: make(map[string]metrics.Metrics),
		logger:  logger,
	}
}

func (storage *MemStorage) UpdateMetric(metricType string, metricName string, value string) error {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	key := metricName
	switch metricType {
	case metrics.Gauge:
		if newValue, err := strconv.ParseFloat(value, 64); err == nil {
			storage.metrics[key] = metrics.Metrics{Value: &newValue}
			storage.logger.Info().Msgf("STORAGE GAUDE UPDATE: %s  %s  %s %v", metricType, metricName, key, newValue)
		} else {
			storage.logger.Error().Msgf("STORAGE GAUDE UPDATE ERROR: %s  %s  %s %v", metricType, metricName, key, newValue)
			return fmt.Errorf("некорректное значение для типа counter: %v", value)

		}

	case metrics.Counter:
		if existingMetric, ok := storage.metrics[key]; ok {
			existingValue := existingMetric.Delta
			newValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				storage.logger.Error().Msgf("STORAGE COUNTER UPDATE ERROR: %s  %s  %s %v", metricType, metricName, key, newValue)

				return fmt.Errorf("некорректное значение для типа counter: %v", value)

			}

			firstSum := *existingValue
			updatedValue := firstSum + newValue
			storage.metrics[key] = metrics.Metrics{Delta: &updatedValue}

			storage.logger.Info().Msgf("STORAGE COUNTER UPDATE: %s  %s  %s %v", metricType, metricName, key, updatedValue)

		} else {
			if newValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				storage.metrics[key] = metrics.Metrics{Delta: &newValue}
			} else {
				storage.logger.Error().Msgf("STORAGE COUNTER UPDATE ERROR: %s  %s  %s %v", metricType, metricName, key, newValue)
				return fmt.Errorf("некорректное значение для типа counter: %v", value)
			}

		}
	default:
		storage.logger.Error().Msgf("STORAGE COUNTER UPDATE ERROR: %s  %s  %s", metricType, metricName, key)
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
