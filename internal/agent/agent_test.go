package agent

import (
	"sync"
	"testing"
	"time"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/api"
)

func BenchmarkSendMetrics(b *testing.B) {

	type args struct {
		counter   *int64
		agentCfg  config.AgentCfg
		results   chan api.Metrics
		chkResult api.MetricsMap
		errCh     chan error
		wg        *sync.WaitGroup
	}

	params := args{
		counter: new(int64),
		agentCfg: config.AgentCfg{
			PollTik: 5,
		},
		results:   make(chan api.Metrics, 32),
		chkResult: api.MetricsMap{Metrics: make(map[string]api.Metrics, 32)},
		errCh:     make(chan error),
		wg:        new(sync.WaitGroup),
	}

	b.Run("pool metric bench", func(b *testing.B) {
		go CollectMetrics(params.counter, params.agentCfg, params.results, params.errCh, params.wg)

		for len(params.chkResult.Metrics) != 32 {
			metric := <-params.results
			params.chkResult.Metrics[metric.ID] = metric
		}

	})
}

func TestSendMetrics(t *testing.T) {
	type args struct {
		metrics  <-chan api.Metrics
		agentCfg config.AgentCfg
		errCh    chan<- error
		rwg      *sync.WaitGroup
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendMetrics(tt.args.metrics, tt.args.agentCfg, tt.args.errCh, tt.args.rwg)
		})
	}
}

func TestCollectMetrics(t *testing.T) {
	type args struct {
		counter    *int64
		agentCfg   config.AgentCfg
		results    chan api.Metrics
		chkResult1 api.MetricsMap
		chkResult2 api.MetricsMap
		errCh      chan error
		wg         *sync.WaitGroup
	}

	params := args{
		counter: new(int64),
		agentCfg: config.AgentCfg{
			PollTik: 1 * time.Microsecond,
		},
		results:    make(chan api.Metrics, 32),
		chkResult1: api.MetricsMap{Metrics: make(map[string]api.Metrics, 32)},
		chkResult2: api.MetricsMap{Metrics: make(map[string]api.Metrics, 32)},
		errCh:      make(chan error),
		wg:         new(sync.WaitGroup),
	}

	t.Run("Change metric test", func(t *testing.T) {
		go CollectMetrics(params.counter, params.agentCfg, params.results, params.errCh, params.wg)

		for len(params.chkResult1.Metrics) != 32 {
			metric := <-params.results
			params.chkResult1.Metrics[metric.ID] = <-params.results
		}

		for len(params.chkResult2.Metrics) != 32 {
			metric := <-params.results
			params.chkResult2.Metrics[metric.ID] = metric
		}

		for nameMetric, metric := range params.chkResult1.Metrics {
			if nameMetric == api.Counter {
				if *metric.Delta == *params.chkResult2.Metrics[nameMetric].Delta {
					t.Errorf("metric %v %v %v is not changed",
						metric.ID, *metric.Delta, *params.chkResult2.Metrics[metric.ID].Delta)
				}
			} else if nameMetric == api.Gauge {
				if *metric.Value == *params.chkResult2.Metrics[metric.ID].Value {
					t.Errorf("metric %v %v %v is not changed",
						metric.ID, *metric.Value, *params.chkResult2.Metrics[metric.ID].Value)
				}
			}
		}
	})
}
