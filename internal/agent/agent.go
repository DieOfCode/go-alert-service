package agent

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	m "github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/rs/zerolog"
)

type Agent interface {
	CollectGaudeMetrics() []m.Metrics
	SendMetric(ctx context.Context, client *http.Client, metrics []m.Metrics, address string) error
}

type MetricsAgent struct {
	//TODO: replace to loger interface
	logger zerolog.Logger
}

func NewMetricsAgent(logger zerolog.Logger) *MetricsAgent {
	return &MetricsAgent{
		logger: logger,
	}
}

func (metricAgent *MetricsAgent) CollectGaudeMetrics() []m.Metrics {
	var collectedMerics []m.Metrics
	var stat runtime.MemStats

	runtime.ReadMemStats(&stat)

	memStatValue := reflect.ValueOf(stat)
	memStatType := memStatValue.Type()

	for _, metricName := range m.GaugeMetrics {

		fieldValue, success := memStatType.FieldByName(metricName)

		if !success {
			continue
		}

		canFloat := memStatValue.FieldByName(metricName).CanFloat()
		if !canFloat {
			continue
		}
		value := memStatValue.FieldByName(metricName).Float()
		collectedMerics = append(collectedMerics, m.Metrics{MType: m.Gauge, ID: fieldValue.Name, Value: &value})

	}

	return collectedMerics
}

func (metricAgent *MetricsAgent) SendMetric(ctx context.Context, client *http.Client, metrics []m.Metrics, address string) error {
	wg := sync.WaitGroup{}

	for _, element := range metrics {
		wg.Add(1)
		go func(element m.Metrics) {
			defer wg.Done()
			metricType := element.MType
			metricID := element.ID
			var request string
			if metricType == m.Gauge {
				if element.Value != nil {
					value := *element.Value
					request = fmt.Sprintf("http://%s/update/%s/%s/%v", address, metricType, metricID, value)
				} else {
					request = fmt.Sprintf("http://%s/update/%s/%s", address, metricType, metricID)
				}

			} else {
				value := *element.Delta
				request = fmt.Sprintf("http://%s/update/%s/%s/%v", address, element.MType, metricID, value)

			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, request, nil)
			if err != nil {
				metricAgent.logger.Err(err).Msgf("REQUEST CREATE ERROR")
				return
			}
			req.Header.Set("Content-Type", "text/plain")
			resp, err := client.Do(req)
			if err != nil {
				metricAgent.logger.Err(err).Msgf("UPDATE METRIC VALUE ERROR %s", element.ID)
				return
			}
			resp.Body.Close()
		}(element)

	}

	wg.Wait()
	return nil
}
