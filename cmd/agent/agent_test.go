package main

import (
	"sync"
	"testing"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/agent"
	"github.com/netzen86/collectmetrics/internal/api"
)

func TestCollectMetrics(t *testing.T) {
	type args struct {
		counter  *int64
		agentCfg config.AgentCfg
		results  chan<- api.Metrics
		errCh    chan<- error
		wg       *sync.WaitGroup
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent.CollectMetrics(tt.args.counter, tt.args.agentCfg, tt.args.results, tt.args.errCh, tt.args.wg)
		})
	}
}
