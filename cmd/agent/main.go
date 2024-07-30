package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

const (
	addressServer string = "http://localhost:8080"
	ct            string = "text/html"
	Alloc
	BuckHashSys
	Frees
	GCCPUFraction
	GCSys
	HeapAlloc
	HeapIdle
	HeapInuse
	HeapObjects
	HeapReleased
	HeapSys
	LastGC
	Lookups
	MCacheInuse
	MCacheSys
	MSpanInuse
	MSpanSys
	Mallocs
	NextGC
	NumForcedGC
	NumGC
	OtherSys
	PauseTotalNs
	StackInuse
	StackSys
	Sys
	TotalAlloc
)

func SendMetrics(url, metricData string) error {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", url, metricData), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", ct)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	return nil
}
func main() {
	runtime.GC()

	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	fmt.Printf("Total allocated memory (in bytes): %d\n", memStats.Alloc)
	fmt.Printf("Heap memory (in bytes): %d\n", memStats.BuckHashSys)
	fmt.Printf("Number of garbage collections: %d\n", memStats.Sys)

}
