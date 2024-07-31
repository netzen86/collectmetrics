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
	runtime.GC()
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)
	fmt.Println(memStats.GCSys)
	storage.UpdateParam(ctx, gag, Alloc, float64(memStats.Alloc))
	storage.UpdateParam(ctx, gag, BuckHashSys, float64(memStats.BuckHashSys))
	storage.UpdateParam(ctx, gag, Frees, float64(memStats.Frees))
	storage.UpdateParam(ctx, gag, GCCPUFraction, float64(memStats.GCCPUFraction))
	storage.UpdateParam(ctx, gag, GCSys, float64(memStats.GCSys))
	storage.UpdateParam(ctx, gag, HeapAlloc, float64(memStats.HeapAlloc))
	storage.UpdateParam(ctx, gag, HeapIdle, float64(memStats.HeapIdle))
	storage.UpdateParam(ctx, gag, HeapInuse, float64(memStats.HeapInuse))
	storage.UpdateParam(ctx, gag, HeapObjects, float64(memStats.HeapObjects))
	storage.UpdateParam(ctx, gag, HeapReleased, float64(memStats.HeapReleased))
	storage.UpdateParam(ctx, gag, HeapSys, float64(memStats.HeapSys))
	storage.UpdateParam(ctx, gag, LastGC, float64(memStats.LastGC))
	storage.UpdateParam(ctx, gag, Lookups, float64(memStats.Lookups))
	storage.UpdateParam(ctx, gag, MCacheInuse, float64(memStats.MCacheInuse))
	storage.UpdateParam(ctx, gag, MCacheSys, float64(memStats.MCacheSys))
	storage.UpdateParam(ctx, gag, MSpanInuse, float64(memStats.MSpanInuse))
	storage.UpdateParam(ctx, gag, Mallocs, float64(memStats.Mallocs))
	storage.UpdateParam(ctx, gag, NextGC, float64(memStats.NextGC))
	storage.UpdateParam(ctx, gag, NumForcedGC, float64(memStats.NumForcedGC))
	storage.UpdateParam(ctx, gag, NumGC, float64(memStats.NumGC))
	storage.UpdateParam(ctx, gag, OtherSys, float64(memStats.OtherSys))
	storage.UpdateParam(ctx, gag, PauseTotalNs, float64(memStats.PauseTotalNs))
	storage.UpdateParam(ctx, gag, StackInuse, float64(memStats.StackInuse))
	storage.UpdateParam(ctx, gag, StackSys, float64(memStats.StackSys))
	storage.UpdateParam(ctx, gag, Sys, float64(memStats.Sys))
	storage.UpdateParam(ctx, gag, TotalAlloc, float64(memStats.TotalAlloc))
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
	fmt.Println(storage.Gauge)

}
