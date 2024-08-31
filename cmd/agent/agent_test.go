package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/netzen86/collectmetrics/internal/db"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

func TestSendMetrics(t *testing.T) {
	type args struct {
		metricData string
	}
	tests := []struct {
		name       string
		args       args
		httpStatus int
		wantErr    bool
	}{
		{
			name: "test count",
			args: args{
				metricData: "/update/counter/PollCount/100",
			},
			httpStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "test guage",
			args: args{
				metricData: "/update/gauge/Alloc/50",
			},
			httpStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "test bad reqest",
			args: args{
				metricData: "/updater/gauge/Alloc/50",
			},
			httpStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
			}))
			if err := SendMetrics(server.URL, tt.args.metricData); (err != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollectMetrics(t *testing.T) {
	storage, err := memstorage.NewMemStorage()
	if err != nil {
		t.Errorf("expected err to be nil got %v", err)
	}
	CollectMetrics(storage, db.DataBaseConString, "MEMORY")
	CollectMetrics(storage, db.DataBaseConString, "MEMORY")
	CollectMetrics(storage, db.DataBaseConString, "MEMORY")
	for key, val := range storage.Gauge {
		if val == 0 && key != "Lookups" {
			t.Errorf("%s not get value", key)
		}
	}
	// deprecated field
	if storage.Gauge["Lookups"] != 0 {
		t.Error("Lookups must be zero")
	}
	if storage.Counter["PollCount"] != 3 {
		t.Error("PollCount not get value")
	}
}
