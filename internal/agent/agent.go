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
)

// TODO replace with real metric type

func CollectGaudeMetrics() []m.Metrics {
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

func SendMetric(ctx context.Context, client *http.Client, metrics []m.Metrics, address string) error {
	wg := sync.WaitGroup{}

	for _, element := range metrics {
		wg.Add(1)
		go func(element m.Metrics) {
			defer wg.Done()
			request := fmt.Sprintf("http://%s/update/%s/%s/%v", address, element.MType, element.ID, element.Value)
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
