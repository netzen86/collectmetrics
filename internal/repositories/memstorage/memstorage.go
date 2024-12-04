package memstorage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/utils"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
	mx      sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Gauge: make(map[string]float64), Counter: make(map[string]int64)}
}

func (storage *MemStorage) UpdateParam(ctx context.Context, cntSummed bool,
	metricType, metricName string, metricValue interface{}, logger zap.SugaredLogger) error {
	switch {
	case metricType == api.Gauge:
		value, err := utils.ParseValGag(metricValue)
		if err != nil {
			return err
		}
		storage.mx.Lock()
		storage.Gauge[metricName] = value
		storage.mx.Unlock()
	case metricType == api.Counter:
		delta, err := utils.ParseValCnt(metricValue)
		if err != nil {
			return err
		}
		if !cntSummed {
			storage.mx.Lock()
			storage.Counter[metricName] += delta
			storage.mx.Unlock()
		} else {
			storage.mx.Lock()
			storage.Counter[metricName] = delta
			storage.mx.Unlock()
		}
	default:
		return errors.New("wrong metric type")
	}

	return nil
}

func (storage *MemStorage) GetAllMetrics(ctx context.Context, logger zap.SugaredLogger) (api.MetricsMap, error) {
	var metrics api.MetricsMap
	metrics.Metrics = make(map[string]api.Metrics)

	for name, value := range storage.Gauge {
		logger.Infoln("METRICS GAUGE WRITE")
		storage.mx.RLock()
		metrics.Metrics[name] = api.Metrics{MType: api.Gauge, ID: name, Value: &value}
		storage.mx.RUnlock()
	}
	for name, delta := range storage.Counter {
		logger.Infoln("METRICS COUNTER WRITE")
		storage.mx.RLock()
		metrics.Metrics[name] = api.Metrics{MType: api.Counter, ID: name, Delta: &delta}
		storage.mx.RUnlock()
	}

	return metrics, nil
}

func (storage *MemStorage) GetCounterMetric(ctx context.Context, metricID string,
	logger zap.SugaredLogger) (int64, error) {
	storage.mx.RLock()
	delta, ok := storage.Counter[metricID]
	storage.mx.RUnlock()
	if !ok {
		return 0, fmt.Errorf("error get counter metric %s", metricID)
	}
	return delta, nil
}

func (storage *MemStorage) GetGaugeMetric(ctx context.Context, metricID string,
	logger zap.SugaredLogger) (float64, error) {
	storage.mx.RLock()
	value, ok := storage.Gauge[metricID]
	storage.mx.RUnlock()
	if !ok {
		return 0, fmt.Errorf("error get gauge metric %s", metricID)
	}
	return value, nil
}

func (storage *MemStorage) CreateTables(ctx context.Context, logger zap.SugaredLogger) error {
	return nil
}
