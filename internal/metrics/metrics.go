package metrics

type MetricType string

const (
	TypeCounter = "counter"
	TypeGauge   = "gauge"
)

type AgentMetric struct {
	MType string `json:"type"`
	ID    string `json:"id"`
	Value any    `json:"value,omitempty"`
	Delta any    `json:"delta,omitempty"`
}

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Error struct {
	Error string `json:"error"`
}

type Data map[string]map[string]Metric

var GaugeMetrics = []string{
	"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
	"TotalAlloc", "Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
	"PauseTotalNs", "StackInuse", "StackSys", "Sys",
	"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
}

var GopsutilGaugeMetrics = []string{
	"TotalMemory",
	"FreeMemory",
	"CPUutilization1",
}
