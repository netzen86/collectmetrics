package memstorage

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/utils"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{Gauge: make(map[string]float64), Counter: make(map[string]int64)}, nil
}

func (storage *MemStorage) UpdateParam(ctx context.Context, cntSummed bool, tempfile, metricType, metricName string, metricValue interface{}) error {
	switch {
	case metricType == "gauge":
		val, err := utils.ParseValGag(metricValue)
		if err != nil {
			return err
		}
		storage.Gauge[metricName] = val

	case metricType == "counter":
		del, err := utils.ParseValCnt(metricValue)
		if err != nil {
			return err
		}
		if !cntSummed {
			storage.Counter[metricName] += del
		} else {
			storage.Counter[metricName] = del
		}
	default:
		return errors.New("wrong metric type")
	}
	return nil
}

func (storage *MemStorage) GetAllMetrics(ctx context.Context) (api.MetricsSlice, error) {
	var metrics api.MetricsSlice

	for k, v := range storage.Gauge {
		log.Println("METRICS GAUGE WRITE")
		metrics.Metrics = append(metrics.Metrics, api.Metrics{MType: api.Gauge, ID: k, Value: &v})
	}
	for k, v := range storage.Counter {
		log.Println("METRICS COUNTER WRITE")
		metrics.Metrics = append(metrics.Metrics, api.Metrics{MType: api.Counter, ID: k, Delta: &v})
	}
	return metrics, nil
}

// функция для получения мемсторожа, param не используется
func (storage *MemStorage) GetStorage(ctx context.Context) (*MemStorage, error) {
	if storage == nil {
		return nil, errors.New("storage not init")
	}
	return storage, nil
}

func (storage *MemStorage) GetCounterMetric(ctx context.Context, metricID string) (int64, error) {
	delta, ok := storage.Counter[metricID]
	if !ok {
		return 0, fmt.Errorf("error get counter metric %s", metricID)
	}
	return delta, nil
}

func (storage *MemStorage) GetGaugeMetric(ctx context.Context, metricID string) (float64, error) {
	value, ok := storage.Gauge[metricID]
	if !ok {
		return 0, fmt.Errorf("error get gauge metric %s", metricID)
	}
	return value, nil
}

func (storage *MemStorage) CreateTables(ctx context.Context) error {
	return nil
}
