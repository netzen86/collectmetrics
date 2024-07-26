package main

import (
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

func getMetrics(storage *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			uri := strings.Split(r.RequestURI, "/")
			// io.WriteString(w, r.RequestURI)
			if len(uri) == 5 {
				err := storage.AddData(uri[2], uri[3], uri[4])
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					w.WriteHeader(http.StatusBadRequest)
				}
				w.WriteHeader(http.StatusOK)
			} else {
				http.Error(w, "Metrics name not found!", http.StatusNotFound)
				w.WriteHeader(http.StatusNotFound)
			}
			return
		} else {
			http.Error(w, "Use POST method!", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func main() {
	memSto := NewMemStorage()
	http.HandleFunc("/", getMetrics(memSto))
	http.Handle("/update", http.NotFoundHandler())
	http.Handle("/update/counter", http.NotFoundHandler())
	http.Handle("/update/gauge", http.NotFoundHandler())

	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
