package memstorage

import (
	"context"
	"errors"
	"strconv"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// type Storage interface {
// 	AddData()
// }

func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{gauge: make(map[string]float64), counter: make(map[string]int64)}, nil
}

func (storage *MemStorage) UpdateParam(ctx context.Context, metricType, metricName, metricValue string) error {
	if metricType == "gauge" {
		mv, convOk := strconv.ParseFloat(metricValue, 64)
		if convOk == nil {
			storage.gauge[metricName] = mv
			return nil
		} else {
			return errors.New("value wrong type")
		}
	}
	if metricType == "counter" {
		mv, convOk := strconv.ParseInt(metricValue, 10, 64)
		if convOk == nil {
			storage.counter[metricName] += mv
			return nil
		} else {
			return errors.New("value wrong type")

		}
	}
	return errors.New("wrong metic type")
}
