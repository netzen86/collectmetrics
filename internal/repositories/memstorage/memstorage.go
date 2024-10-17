package memstorage

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/utils"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Gauge: make(map[string]float64), Counter: make(map[string]int64)}
}

func (storage *MemStorage) UpdateParam(ctx context.Context, cntSummed bool,
	metricType, metricName string, metricValue interface{}, logger zap.SugaredLogger) error {
	if metricType == api.Gauge {
		value, err := utils.ParseValGag(metricValue)
		if err != nil {
			return err
		}
		storage.Gauge[metricName] = value
	} else if metricType == api.Counter {
		delta, err := utils.ParseValCnt(metricValue)
		if err != nil {
			return err
		}
		if !cntSummed {
			storage.Counter[metricName] += delta
		} else {
			storage.Counter[metricName] = delta
		}
	} else {
		return errors.New("wrong metric type")
	}
	return nil
}

func (storage *MemStorage) GetAllMetrics(ctx context.Context, logger zap.SugaredLogger) (api.MetricsMap, error) {
	var metrics api.MetricsMap

	metrics.Metrics = make(map[string]api.Metrics)

	for name, value := range storage.Gauge {
		logger.Infoln("METRICS GAUGE WRITE")
		metrics.Metrics[name] = api.Metrics{MType: api.Gauge, ID: name, Value: &value}
	}
	for name, delta := range storage.Counter {
		logger.Infoln("METRICS COUNTER WRITE")
		metrics.Metrics[name] = api.Metrics{MType: api.Counter, ID: name, Delta: &delta}
	}
	return metrics, nil
}

func (storage *MemStorage) GetCounterMetric(ctx context.Context, metricID string,
	logger zap.SugaredLogger) (int64, error) {
	delta, ok := storage.Counter[metricID]
	if !ok {
		return 0, fmt.Errorf("error get counter metric %s", metricID)
	}
	return delta, nil
}

func (storage *MemStorage) GetGaugeMetric(ctx context.Context, metricID string,
	logger zap.SugaredLogger) (float64, error) {
	value, ok := storage.Gauge[metricID]
	if !ok {
		return 0, fmt.Errorf("error get gauge metric %s", metricID)
	}
	return value, nil
}

func (storage *MemStorage) CreateTables(ctx context.Context, logger zap.SugaredLogger) error {
	return nil
}

func MetricMapToMemstorage(metrics *api.MetricsMap, storage MemStorage) {
	for _, metric := range metrics.Metrics {
		if metric.MType == api.Gauge {
			storage.Gauge[metric.ID] = *metric.Value
		}
		if metric.MType == api.Counter {
			storage.Counter[metric.ID] = *metric.Delta
		}
	}
}

func MemstoragetoMetricMap(metrics *api.MetricsMap, storage MemStorage) {
	for name, value := range storage.Gauge {
		metrics.Metrics[name] = api.Metrics{ID: name, MType: api.Gauge, Value: &value}
	}
	for name, delta := range storage.Counter {
		metrics.Metrics[name] = api.Metrics{ID: name, MType: api.Counter, Delta: &delta}
	}
}
