package memstorage

import (
	"context"
	"errors"

	"github.com/netzen86/collectmetrics/internal/utils"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{Gauge: make(map[string]float64), Counter: make(map[string]int64)}, nil
}

func (storage *MemStorage) UpdateParam(ctx context.Context, cntSummed bool, metricType, metricName string, metricValue interface{}) error {
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

func (storage *MemStorage) GetMemStorage(ctx context.Context) (*MemStorage, error) {
	if storage == nil {
		return nil, errors.New("storage empty")
	}
	return storage, nil
}
