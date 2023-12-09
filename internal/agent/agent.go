package agent

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	m "github.com/DieOfCode/go-alert-service/internal/metrics"
)

// TODO replace with real metric type

func CollectGaudeMetrics() []m.Metric {
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

func SendMetric(ctx context.Context, client http.Client, metrics []m.Metric) error {
	wg := sync.WaitGroup{}
	wg.Add(len(metrics))

	for _, element := range metrics {

		go func(element m.Metric) {
			defer wg.Done()
			request := fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", element.MetricType, element.MetricName, element.Value)
			req, err := http.NewRequestWithContext(ctx, "POST", request, nil)
			if err != nil {
				//TODO handle error
				return
			}
			req.Header.Set("Content-Type", "text/plain")
			resp, err := client.Do(req)
			if err != nil {
				//TODO handle error
				return
			}
			resp.Body.Close()
		}(element)

	}

	wg.Wait()
	return nil
}
