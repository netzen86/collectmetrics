package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"errors"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

type Storage interface {
	AddData()
}

func NewMemStorage() *MemStorage {
	return &MemStorage{gauge: make(map[string]float64), counter: make(map[string]int64)}
}

func (storage *MemStorage) AddData(metricType, metricName string, metricValue interface{}) error {
	if metricType == "gauge" {
		mv, convOk := metricValue.(float64)
		if convOk {
			storage.gauge[metricName] = mv
		} else {
			return errors.New("value wrong type")
		}
	}
	if metricType == "counter" {
		mv, convOk := metricValue.(int64)
		if convOk {
			_, ok := storage.counter[metricName]
			if !ok {
				storage.counter[metricName] = mv
			} else {
				storage.counter[metricName] += mv
			}
		} else {
			return errors.New("value wrong type")

		}
	}
	return nil
}

func getMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		storage := NewMemStorage()
		uri := strings.Split(r.RequestURI, "/")
		io.WriteString(w, r.RequestURI)
		if len(uri) == 4 {
			storage.AddData(uri[1], uri[2], strconv.Atoi(uri[3]))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		fmt.Println(storage)
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	err := http.ListenAndServe(`:8080`, http.HandlerFunc(getMetrics))
	if err != nil {
		panic(err)
	}
}
