package metrics

import "reflect"

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

const (
	PoolCount   string = "PollCount"
	RandomValue string = "RandomValue"
)

type Metric struct {
	MetricType MetricType
	MetricName string
	Value      interface{}
}

func (metric *Metric) ToMetrics() Metrics {
	id := metric.MetricName
	v := metric.Value

	if reflect.TypeOf(v).Kind() == reflect.Int64 {
		// Extract the underlying int64 value
		int64Value := reflect.ValueOf(v).Int()
		print(int64Value)
		if metric.MetricType == "gauge" {
			value := metric.Value.(float64)
			return Metrics{ID: id, Value: &value, MType: "gauge"}

		} else {
			value := metric.Value.(int64)
			return Metrics{ID: id, Delta: &value, MType: "counter"}
		}
	} else {
		if metric.MetricType == "gauge" {
			value := metric.Value.(float64)
			return Metrics{ID: id, Value: &value, MType: "gauge"}

		} else {
			value := metric.Value.(int64)
			return Metrics{ID: id, Delta: &value, MType: "counter"}
		}
	}
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var GaugeMetrics = []string{
	"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
	"TotalAlloc", "Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
	"PauseTotalNs", "StackInuse", "StackSys", "Sys",
	"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
}
