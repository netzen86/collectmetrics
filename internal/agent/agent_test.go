package agent

import (
	"sync"
	"testing"
	"time"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/logger"
	"github.com/stretchr/testify/assert"
)

func BenchmarkSendMetrics(b *testing.B) {

	testLogger, err := logger.Logger()
	if err != nil {
		b.Errorf("error when get agent logger %v", err)
	}

	type args struct {
		counter   *int64
		results   chan api.Metrics
		chkResult api.MetricsMap
		errCh     chan error
		wg        *sync.WaitGroup
		agentCfg  config.AgentCfg
	}

	params := args{
		counter: new(int64),
		agentCfg: config.AgentCfg{
			PollTik: 1 * time.Millisecond,
			Logger:  testLogger,
		},
		results:   make(chan api.Metrics, 32),
		chkResult: api.MetricsMap{Metrics: make(map[string]api.Metrics, 32)},
		errCh:     make(chan error),
		wg:        new(sync.WaitGroup),
	}

	b.Run("pool metric bench", func(b *testing.B) {
		go CollectMetrics(params.counter, params.agentCfg, params.results, params.errCh, params.wg)

		for len(params.chkResult.Metrics) < 31 {
			metric := <-params.results
			params.chkResult.Metrics[metric.ID] = metric
		}

	})
}

func TestSendMetrics(t *testing.T) {
	type args struct {
		metrics  <-chan api.Metrics
		errCh    chan<- error
		rwg      *sync.WaitGroup
		agentCfg config.AgentCfg
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
	testLogger, err := logger.Logger()
	if err != nil {
		t.Errorf("error when get agent logger %v", err)
	}
	// defer func() {
	// 	err = testLogger.Sync()
	// 	if err != nil {
	// 		t.Errorf("error when sync logger %v", err)
	// 	}
	// }()

	type args struct {
		counter    *int64
		results    chan api.Metrics
		chkResult1 api.MetricsMap
		chkResult2 api.MetricsMap
		errCh      chan error
		wg         *sync.WaitGroup
		agentCfg   config.AgentCfg
	}

	params := args{
		counter: new(int64),
		agentCfg: config.AgentCfg{
			PollTik: 1 * time.Millisecond,
			Logger:  testLogger,
		},
		results:    make(chan api.Metrics, 32),
		chkResult1: api.MetricsMap{Metrics: make(map[string]api.Metrics, 32)},
		chkResult2: api.MetricsMap{Metrics: make(map[string]api.Metrics, 32)},
		errCh:      make(chan error),
		wg:         new(sync.WaitGroup),
	}

	go CollectMetrics(params.counter, params.agentCfg, params.results, params.errCh, params.wg)

	for len(params.chkResult1.Metrics) < 32 {
		metric := <-params.results
		params.chkResult1.Metrics[metric.ID] = metric
	}

	for len(params.chkResult2.Metrics) < 32 {
		metric := <-params.results
		params.chkResult2.Metrics[metric.ID] = metric
	}

	t.Run("Changed metric PollCount ", func(t *testing.T) {
		assert.NotEqual(t,
			*params.chkResult1.Metrics[config.PollCount].Delta,
			*params.chkResult2.Metrics[config.PollCount].Delta,
			"PoolCount not changed")
	})

	t.Run("Changed metric RandomValue ", func(t *testing.T) {
		assert.NotEqual(t,
			*params.chkResult1.Metrics[config.RandomValue].Value,
			*params.chkResult2.Metrics[config.RandomValue].Value,
			"RandomValue not changed")
	})

}

func ExampleCollectMetrics() {
	// инициализация логгера
	testLogger, err := logger.Logger()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = testLogger.Sync()
	}()

	// оъявляем структуру с полями необходимыми для работы функции CollectMetrics
	agentCfg := config.AgentCfg{
		PollTik: 1 * time.Millisecond,
		Logger:  testLogger,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	// запускаем функцию CollectMetrics
	go CollectMetrics(new(int64), agentCfg, make(chan api.Metrics, 32), make(chan error), &wg)
}
