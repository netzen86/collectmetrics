package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/spf13/pflag"
)

const (
	addressServer  string        = "localhost:8080"
	ct             string        = "text/html"
	gag            string        = "gauge"
	cnt            string        = "counter"
	Alloc          string        = "Alloc"
	BuckHashSys    string        = "BuckHashSys"
	Frees          string        = "Frees"
	GCCPUFraction  string        = "GCCPUFraction"
	GCSys          string        = "GCSys"
	HeapAlloc      string        = "HeapAlloc"
	HeapIdle       string        = "HeapIdle"
	HeapInuse      string        = "HeapInuse"
	HeapObjects    string        = "HeapObjects"
	HeapReleased   string        = "HeapReleased"
	HeapSys        string        = "HeapSys"
	LastGC         string        = "LastGC"
	Lookups        string        = "Lookups"
	MCacheInuse    string        = "MCacheInuse"
	MCacheSys      string        = "MCacheSys"
	MSpanInuse     string        = "MSpanInuse"
	MSpanSys       string        = "MSpanSys"
	Mallocs        string        = "Mallocs"
	NextGC         string        = "NextGC"
	NumForcedGC    string        = "NumForcedGC"
	NumGC          string        = "NumGC"
	OtherSys       string        = "OtherSys"
	PauseTotalNs   string        = "PauseTotalNs"
	StackInuse     string        = "StackInuse"
	StackSys       string        = "StackSys"
	Sys            string        = "Sys"
	TotalAlloc     string        = "TotalAlloc"
	PollCount      string        = "PollCount"
	RandomValue    string        = "RandomValue"
	pollInterval   time.Duration = 2 * time.Second
	reportInterval time.Duration = 10 * time.Second
)

func CollectMetrics(storage repositories.Repo) {
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
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	return nil
}
func main() {
	var endpoint string
	var pInterv time.Duration
	var rInterv time.Duration
	pflag.StringVarP(&endpoint, "endpoint", "a", addressServer, "Used to set the address and port to connect server.")
	pflag.DurationVarP(&pInterv, "pollinterval", "p", pollInterval, "User for set poll interval in seconds.")
	pflag.DurationVarP(&rInterv, "reportinterval", "r", reportInterval, "User for set report interval (send to srv) in seconds.")
	pflag.Parse()
	if len(pflag.Args()) != 0 {
		pflag.PrintDefaults()
		os.Exit(1)
	}
	pollTik := time.NewTicker(pInterv)
	reportTik := time.NewTicker(rInterv)

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
				SendMetrics(fmt.Sprintf("http://%s/update/", endpoint), fmt.Sprintf("gauge/%s/%v", k, v))
			}
			for k, v := range storage.Counter {
				SendMetrics(fmt.Sprintf("http://%s/update/", endpoint), fmt.Sprintf("counter/%s/%v", k, v))
			}
		}
	}
}
