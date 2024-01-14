package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"sync"

	m "github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/rs/zerolog"
)

type Agent interface {
	CollectGaudeMetrics() []m.Metric
	SendMetric(ctx context.Context, client *http.Client, metrics []m.Metric, address string, pool *sync.Pool) error
}

type MetricAgent struct {
	logger zerolog.Logger
}

func NewMetricAgent(logger zerolog.Logger) *MetricAgent {
	return &MetricAgent{
		logger: logger,
	}
}
func (agent *MetricAgent) CollectGaudeMetrics() []m.Metric {
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

func (agent *MetricAgent) SendMetric(ctx context.Context, client *http.Client, metrics []m.Metric, address string, pool *sync.Pool) error {
	wg := sync.WaitGroup{}

	for _, element := range metrics {
		wg.Add(1)
		go func(element m.Metric) {
			b, err := json.Marshal(element.ToMetrics())
			if err != nil {
				return
			}
			defer wg.Done()
			buf := &bytes.Buffer{}
			gw := pool.Get().(*gzip.Writer)
			defer pool.Put(gw)
			gw.Reset(buf)
			agent.logger.Info().Msgf("buffer points to: %p", buf)
			agent.logger.Info().Msgf("buffer's content: %v", buf.String())
			n, err := gw.Write(b)
			if err != nil {
				return
			}
			agent.logger.Info().Msgf("buffer's content: %v", buf.String())
			gw.Close()

			agent.logger.Info().
				Int("len of b", len(b)).
				Int("written bytes", n).
				Int("len of buf", len(buf.Bytes())).
				Send()

			request := fmt.Sprintf("http://%s/update/", address)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, request, buf)
			if err != nil {
				agent.logger.Err(err)
				return
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Encoding", "gzip")
			resp, err := client.Do(req)
			if err != nil {
				agent.logger.Info().Msgf("buffer's content: %v", buf.String())
				log.Println(err)
				return
			}
			resp.Body.Close()
		}(element)

	}

	wg.Wait()
	return nil
}
