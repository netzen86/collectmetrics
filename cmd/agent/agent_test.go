package main

import (
	"testing"

	"github.com/netzen86/collectmetrics/internal/agent"
	"github.com/netzen86/collectmetrics/internal/api"
)

func TestCollectMetrics(t *testing.T) {
	type args struct {
		counter *int64
		numJobs int
		results chan api.Metrics
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent.CollectMetrics(tt.args.counter, tt.args.numJobs, tt.args.results)
		})
	}
}
