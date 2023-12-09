package handler

import (
	"net/http"
	"testing"

	s "github.com/DieOfCode/go-alert-service/internal/storage"
)

func TestHandler_HandleUpdateMetric(t *testing.T) {
	type fields struct {
		repository s.Repository
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Handler{
				repository: tt.fields.repository,
			}
			m.HandleUpdateMetric(tt.args.w, tt.args.r)
		})
	}
}
