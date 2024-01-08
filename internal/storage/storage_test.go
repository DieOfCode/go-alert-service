package storage

// import (
// 	"sync"
// 	"testing"

// 	"github.com/DieOfCode/go-alert-service/internal/metrics"
// )

// func TestMemStorage_UpdateMetric(t *testing.T) {
// 	type fields struct {
// 		mu      sync.Mutex
// 		metrics map[string]metrics.Metric
// 	}
// 	type args struct {
// 		metricType metrics.MetricType
// 		metricName string
// 		value      string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			m := &MemStorage{
// 				mu:      tt.fields.mu,
// 				metrics: tt.fields.metrics,
// 			}
// 			if err := m.UpdateMetric(tt.args.metricType, tt.args.metricName, tt.args.value); (err != nil) != tt.wantErr {
// 				t.Errorf("MemStorage.UpdateMetric() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
