package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	m "github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/rs/zerolog"
)

// TODO replace with real metric type

type Agent struct {
	logger zerolog.Logger
}

type MetricAgent interface {
	CollectGaudeMetrics() []m.Metric
	SendMetric(ctx context.Context, client *http.Client, metrics []m.Metric, address string) error
}

func NewAgent(logger zerolog.Logger) *Agent {
	return &Agent{
		logger: logger,
	}
}

func (agent *Agent) CollectGaudeMetrics() []m.Metric {
	var collectedMerics []m.Metric
	var stat runtime.MemStats

	runtime.ReadMemStats(&stat)

	memStatValue := reflect.ValueOf(stat)
	memStatType := memStatValue.Type()

	for _, metricName := range m.GaugeMetrics {

		fieldValue, success := memStatType.FieldByName(metricName)

		if !success {
			continue
		}

		value := memStatValue.FieldByName(metricName)

		collectedMerics = append(collectedMerics, m.Metric{MetricType: m.Gauge, MetricName: fieldValue.Name, Value: value})

	}

	return collectedMerics
}

func (agent *Agent) SendMetric(ctx context.Context, client *http.Client, metrics []m.Metric, address string) error {
	wg := sync.WaitGroup{}

	for _, element := range metrics {
		wg.Add(1)
		go func(element m.Metric) {
			defer wg.Done()
			request := fmt.Sprintf("http://%s/update/%s/%s/%v", address, element.MetricType, element.MetricName, element.Value)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, request, nil)
			if err != nil {
				log.Println(err)
				return
			}
			req.Header.Set("Content-Type", "text/plain")
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			resp.Body.Close()
		}(element)

	}

	wg.Wait()
	return nil
}
