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

func NewMemStorage() (*MemStorage, error) {
	return &MemStorage{gauge: make(map[string]float64), counter: make(map[string]int64)}, nil
}

func (storage *MemStorage) UpdateParam(ctx context.Context, metricType, metricName string, metricValue interface{}) error {
	if metricType == "gauge" {
		var mv float64
		var convOk error
		switch metricValue.(type) {
		case string:
			mv, convOk = strconv.ParseFloat(metricValue.(string), 64)
			if convOk != nil {
				return errors.New("value wrong type")
			}
		case float64:
			mv = metricValue.(float64)
		default:
			return errors.New("value wrong type")
		}
		storage.gauge[metricName] = mv
	}
	if metricType == "counter" {
		var mv int64
		var convOk error
		switch metricValue.(type) {
		case string:
			mv, convOk = strconv.ParseInt(metricValue.(string), 10, 64)
			if convOk != nil {
				return errors.New("value wrong type")
			}
		case int64:
			mv = metricValue.(int64)

		}
		storage.counter[metricName] += mv
	}
	return nil
}
