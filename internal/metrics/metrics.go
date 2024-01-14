package metrics

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
	MetricType MetricType `json:"id"`
	MetricName string     `json:"type"`
	Value      *float64   `json:"value,omitempty"`
	Delta      *int64     `json:"delta,omitempty"`
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
