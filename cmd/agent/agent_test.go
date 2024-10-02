package main

import (
	"testing"

	"github.com/netzen86/collectmetrics/internal/agent"
)

func TestCollectMetrics(t *testing.T) {
	type args struct {
		counter *int64
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent.CollectMetrics(tt.args.counter)
		})
	}
}
