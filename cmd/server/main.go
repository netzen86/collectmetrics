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

func (storage *MemStorage) AddData(metricType, metricName string, metricValue string) error {
	if metricType == "gauge" {
		mv, convOk := strconv.ParseFloat(metricValue, 64)
		// mv, convOk := metricValue.(float64)
		if convOk == nil {
			storage.gauge[metricName] = mv
		} else {
			return errors.New("value wrong type")
		}
	}
	if metricType == "counter" {
		mv, convOk := strconv.ParseInt(metricValue, 10, 64)
		// mv, convOk := metricValue.(int64)
		if convOk == nil {
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

func getMetrics(storage *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			uri := strings.Split(r.RequestURI, "/")
			io.WriteString(w, r.RequestURI)
			if len(uri) == 5 {
				storage.AddData(uri[2], uri[3], uri[4])
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
			fmt.Println("*****", storage)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func main() {
	memSto := NewMemStorage()
	http.HandleFunc("/", getMetrics(memSto))
	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
