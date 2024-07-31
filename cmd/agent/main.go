package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

const (
	addressServer string = "http://localhost:8080"
	ct                   = "text/html"
	gag                  = "gauge"
	cnt                  = "counter"
	Alloc                = "Alloc"
	BuckHashSys          = "BuckHashSys"
	Frees                = "Frees"
	GCCPUFraction        = "GCCPUFraction"
	GCSys                = "GCSys"
	HeapAlloc            = "HeapAlloc"
	HeapIdle             = "HeapIdle"
	HeapInuse            = "HeapInuse"
	HeapObjects          = "HeapObjects"
	HeapReleased         = "HeapReleased"
	HeapSys              = "HeapSys"
	LastGC               = "LastGC"
	Lookups              = "Lookups"
	MCacheInuse          = "MCacheInuse"
	MCacheSys            = "MCacheSys"
	MSpanInuse           = "MSpanInuse"
	MSpanSys             = "MSpanSys"
	Mallocs              = "Mallocs"
	NextGC               = "NextGC"
	NumForcedGC          = "NumForcedGC"
	NumGC                = "NumGC"
	OtherSys             = "OtherSys"
	PauseTotalNs         = "PauseTotalNs"
	StackInuse           = "StackInuse"
	StackSys             = "StackSys"
	Sys                  = "Sys"
	TotalAlloc           = "TotalAlloc"
)

func CollectMetrics(storage *memstorage.MemStorage) {
	ctx := context.Background()
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)
	storage.UpdateParam(ctx, gag, Alloc, float64(memStats.Alloc))
}

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
	storage, err := memstorage.NewMemStorage()
	if err != nil {
		panic("couldnt alloc mem")
	}
	CollectMetrics(storage)
	fmt.Println(*storage)

}
