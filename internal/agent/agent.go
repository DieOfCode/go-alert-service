package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DieOfCode/go-alert-service/internal/configuration"
	m "github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/mem"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Agent struct {
	mu      sync.Mutex
	logger  *zerolog.Logger
	client  HTTPClient
	Metrics []m.AgentMetric
	address string
	key     string
	counter *int64
	gw      *gzip.Writer
}

func New(logger *zerolog.Logger, client HTTPClient, config *configuration.Config) *Agent {
	counter := new(int64)
	*counter = 0
	return &Agent{
		logger:  logger,
		client:  client,
		address: config.ServerAddress,
		counter: counter,
		key:     config.Key,
		gw:      gzip.NewWriter(io.Discard),
		Metrics: make([]m.AgentMetric, len(m.GaugeMetrics)+5),
	}
}

// func (a *Agent) SendMetrics(ctx context.Context) error {
// 	b, err := json.Marshal(a.Metrics)
// 	if err != nil {
// 		a.logger.Error().Err(err).Msg("Marshalling error")
// 		return err
// 	}
// 	a.logger.Info().Any("json", string(b)).Msg("Marshalled")
// 	buf := &bytes.Buffer{}
// 	a.gw.Reset(buf)
// 	n, err := a.gw.Write(b)
// 	if err != nil {
// 		a.logger.Error().Err(err).Msg("gw.Write error")
// 		return err
// 	}
// 	a.gw.Close()
// 	a.logger.Info().
// 		Int("len of b", len(b)).
// 		Int("written bytes", n).
// 		Int("len of buf", len(buf.Bytes())).
// 		Send()
// 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/updates/", a.address), buf)
// 	if err != nil {
// 		a.logger.Error().Err(err).Msg("http.NewRequestWithContext method error")
// 		return err
// 	}
// 	if a.key != "" {
// 		buf2 := *buf
// 		h := hmac.New(sha256.New, []byte(a.key))
// 		if _, err := h.Write(buf2.Bytes()); err != nil {
// 			return err
// 		}
// 		d := h.Sum(nil)
// 		a.logger.Info().Msgf("hash: %x", d)
// 		req.Header.Add("HashSHA256", hex.EncodeToString(d))
// 	}
// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Content-Encoding", "gzip")
// 	res, err := a.client.Do(req)
// 	if err != nil {
// 		a.logger.Error().Err(err).Msg("client.Do method error")
// 		return err
// 	}
// 	res.Body.Close()
// 	a.logger.Info().Any("metric", a.Metrics).Msg("Metrics are sent")
// 	return nil
// }

func (a *Agent) PrepareMetrics(ctx context.Context, interval time.Duration) <-chan []m.AgentMetric {
	ch := make(chan []m.AgentMetric)
	wg := &sync.WaitGroup{}

	poll := time.NewTicker(interval)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-poll.C:
				ch <- a.Metrics
			case <-ctx.Done():
				poll.Stop()
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func (a *Agent) SendMetrics(ctx context.Context, metrics <-chan []m.AgentMetric) error {
	for {
		m, ok := <-metrics
		if !ok {
			return nil
		} else {
			b, err := json.Marshal(m)
			if err != nil {
				a.logger.Error().Err(err).Msg("Marshalling error")
				return err
			}
			a.logger.Info().Any("json", string(b)).Msg("Marshalled")
			buf := &bytes.Buffer{}
			a.gw.Reset(buf)
			n, err := a.gw.Write(b)
			if err != nil {
				a.logger.Error().Err(err).Msg("gw.Write error")
				return err
			}
			a.gw.Close()
			a.logger.Info().
				Int("len of b", len(b)).
				Int("written bytes", n).
				Int("len of buf", len(buf.Bytes())).
				Send()
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/updates/", a.address), buf)
			if err != nil {
				a.logger.Error().Err(err).Msg("http.NewRequestWithContext method error")
				return err
			}
			if a.key != "" {
				buf2 := *buf
				h := hmac.New(sha256.New, []byte(a.key))
				if _, err := h.Write(buf2.Bytes()); err != nil {
					return err
				}
				d := h.Sum(nil)
				a.logger.Info().Msgf("hash: %x", d)
				req.Header.Add("HashSHA256", hex.EncodeToString(d))
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Encoding", "gzip")
			res, err := a.client.Do(req)
			if err != nil {
				a.logger.Error().Err(err).Msg("client.Do method error")
				return err
			}
			res.Body.Close()
			a.logger.Info().Any("metric", a.Metrics).Msg("Metrics are sent")
			return nil
		}
	}
}

func (a *Agent) CollectRuntimeMetrics(ctx context.Context, interval time.Duration) {
	poll := time.NewTicker(interval)

	for {
		select {
		case <-poll.C:
			atomic.AddInt64(a.counter, 1)
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			msvalue := reflect.ValueOf(memStats)
			mstype := msvalue.Type()

			for _, metric := range m.GaugeMetrics {
				field, ok := mstype.FieldByName(metric)
				if !ok {
					return
				}
				value := msvalue.FieldByName(metric).Interface()
				a.Metrics = append(a.Metrics, m.AgentMetric{MType: m.TypeGauge, ID: field.Name, Value: value})
			}

			a.Metrics = append(a.Metrics, m.AgentMetric{MType: m.TypeGauge, ID: "RandomValue", Value: rand.Float64()})
			a.Metrics = append(a.Metrics, m.AgentMetric{MType: m.TypeCounter, ID: "PollCount", Delta: *a.counter})
		case <-ctx.Done():
			poll.Stop()
			return
		}
	}
}

func (a *Agent) CollectGopsutilMetrics(ctx context.Context, interval time.Duration) {
	poll := time.NewTicker(interval)

	for {
		select {
		case <-poll.C:
			v, err := mem.VirtualMemory()
			if err != nil {
				return
			}
			a.Metrics = append(a.Metrics, m.AgentMetric{MType: m.TypeGauge, ID: "TotalMemory", Value: int64(v.Total)})
			a.Metrics = append(a.Metrics, m.AgentMetric{MType: m.TypeGauge, ID: "FreeMemory", Value: int64(v.Free)})
			a.Metrics = append(a.Metrics, m.AgentMetric{MType: m.TypeGauge, ID: "CPUutilization1", Value: v.UsedPercent})
		case <-ctx.Done():
			poll.Stop()
			return
		}
	}
}

func (a *Agent) SendAllMetrics(ctx context.Context) error {
	b, err := json.Marshal(a.Metrics)
	if err != nil {
		a.logger.Error().Err(err).Msg("Marshalling error")
		return err
	}
	a.logger.Info().Any("json", string(b)).Msg("Marshalled")

	buf := &bytes.Buffer{}
	a.gw.Reset(buf)
	n, err := a.gw.Write(b)
	if err != nil {
		a.logger.Error().Err(err).Msg("gw.Write error")
		return err
	}
	a.gw.Close()

	a.logger.Info().
		Int("len of b", len(b)).
		Int("written bytes", n).
		Int("len of buf", len(buf.Bytes())).
		Send()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/updates/", a.address), buf)
	if err != nil {
		a.logger.Error().Err(err).Msg("http.NewRequestWithContext method error")
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")

	res, err := a.client.Do(req)
	if err != nil {
		a.logger.Error().Err(err).Msg("client.Do method error")
		return err
	}
	res.Body.Close()
	a.logger.Info().Any("metric", a.Metrics).Msg("Metrics are sent")
	return nil
}

func (a *Agent) Retry(ctx context.Context, maxRetries int, fn func(ctx context.Context) error) error {
	// Инициализация экспоненциальной стратегии отката
	expBackOff := backoff.NewExponentialBackOff()
	expBackOff.MaxElapsedTime = 0                               // Убираем ограничение по времени
	expBackOff.MaxInterval = backoff.DefaultMaxInterval         // Максимальный интервал между попытками
	expBackOff.InitialInterval = backoff.DefaultInitialInterval // Начальный интервал

	operation := func() error {
		err := fn(ctx)
		if err != nil {
			a.logger.Info().Msg("Retrying...")
		}
		return err
	}

	// Воспользуемся функцией Retry из библиотеки backoff
	err := backoff.Retry(operation, backoff.WithMaxRetries(expBackOff, uint64(maxRetries)))
	if err != nil {
		a.logger.Error().Err(err).Msg("Retrying... Failed")
	}
	return err
}
