package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

const (
	addressServer  string = "http://localhost:8080/update/"
	ct                    = "text/html"
	gag                   = "gauge"
	cnt                   = "counter"
	Alloc                 = "Alloc"
	BuckHashSys           = "BuckHashSys"
	Frees                 = "Frees"
	GCCPUFraction         = "GCCPUFraction"
	GCSys                 = "GCSys"
	HeapAlloc             = "HeapAlloc"
	HeapIdle              = "HeapIdle"
	HeapInuse             = "HeapInuse"
	HeapObjects           = "HeapObjects"
	HeapReleased          = "HeapReleased"
	HeapSys               = "HeapSys"
	LastGC                = "LastGC"
	Lookups               = "Lookups"
	MCacheInuse           = "MCacheInuse"
	MCacheSys             = "MCacheSys"
	MSpanInuse            = "MSpanInuse"
	MSpanSys              = "MSpanSys"
	Mallocs               = "Mallocs"
	NextGC                = "NextGC"
	NumForcedGC           = "NumForcedGC"
	NumGC                 = "NumGC"
	OtherSys              = "OtherSys"
	PauseTotalNs          = "PauseTotalNs"
	StackInuse            = "StackInuse"
	StackSys              = "StackSys"
	Sys                   = "Sys"
	TotalAlloc            = "TotalAlloc"
	PollCount             = "PollCount"
	RandomValue           = "RandomValue"
	pollInterval          = 2 * time.Second
	reportInterval        = 10 * time.Second
)

func CollectMetrics(storage *memstorage.MemStorage) {
	ctx := context.Background()
	var memStats runtime.MemStats

	runtime.GC()

	runtime.ReadMemStats(&memStats)
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
	storage.UpdateParam(ctx, gag, RandomValue, rand.Float64())
	storage.UpdateParam(ctx, cnt, PollCount, int64(1))
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
	response.Body.Close()
	return nil
}
func main() {
	pollTik := time.NewTicker(pollInterval)
	reportTik := time.NewTicker(reportInterval)

	storage, err := memstorage.NewMemStorage()
	if err != nil {
		panic("couldn't alloc mem")
	}
	for {
		select {
		case <-pollTik.C:
			CollectMetrics(storage)
		case <-reportTik.C:
			for k, v := range storage.Gauge {
				SendMetrics(addressServer, fmt.Sprintf("gauge/%s/%v", k, v))
			}
			for k, v := range storage.Counter {
				SendMetrics(addressServer, fmt.Sprintf("counter/%s/%v", k, v))
			}
		}
	}
}
