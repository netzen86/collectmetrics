package main

import (
	"testing"

	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

func TestSendMetrics(t *testing.T) {
	type args struct {
		url        string
		metricData string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMetrics(tt.args.url, tt.args.metricData); (err != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollectMetrics(t *testing.T) {
	type args struct {
		storage *memstorage.MemStorage
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CollectMetrics(tt.args.storage)
		})
	}
}
