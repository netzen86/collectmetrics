package memstorage

import (
	"context"
	"errors"
	"strconv"
)

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{Gauge: make(map[string]float64), Counter: make(map[string]int64)}, nil
}

func (storage *MemStorage) UpdateParam(ctx context.Context, metricType, metricName string, metricValue interface{}) error {
	switch {
	case metricType == "gauge":
		var mv float64
		var convOk error
		switch metricValue := metricValue.(type) {
		case string:
			mv, convOk = strconv.ParseFloat(metricValue, 64)
			if convOk != nil {
				return errors.New("value wrong type")
			}
		case float64:
			mv = metricValue
		default:
			return errors.New("value wrong type")
		}
		storage.Gauge[metricName] = mv

	case metricType == "counter":
		var mv int64
		var convOk error
		switch metricValue := metricValue.(type) {
		case string:
			mv, convOk = strconv.ParseInt(metricValue, 10, 64)
			if convOk != nil {
				return errors.New("value wrong type")
			}
		case int64:
			mv = metricValue

		}
		storage.Counter[metricName] += mv
	default:
		return errors.New("wrong metric type")
	}
	return nil
}

func (storage *MemStorage) GetMemStorage(ctx context.Context) (*MemStorage, error) {
	if storage == nil {
		return nil, errors.New("Storage empty")
	}
	return storage, nil
}
