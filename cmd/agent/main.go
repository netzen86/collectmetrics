package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/spf13/pflag"
)

const (
	addressServer  string = "localhost:8080"
	ct             string = "text/html"
	js             string = "application/json"
	gag            string = "gauge"
	cnt            string = "counter"
	Alloc          string = "Alloc"
	BuckHashSys    string = "BuckHashSys"
	Frees          string = "Frees"
	GCCPUFraction  string = "GCCPUFraction"
	GCSys          string = "GCSys"
	HeapAlloc      string = "HeapAlloc"
	HeapIdle       string = "HeapIdle"
	HeapInuse      string = "HeapInuse"
	HeapObjects    string = "HeapObjects"
	HeapReleased   string = "HeapReleased"
	HeapSys        string = "HeapSys"
	LastGC         string = "LastGC"
	Lookups        string = "Lookups"
	MCacheInuse    string = "MCacheInuse"
	MCacheSys      string = "MCacheSys"
	MSpanInuse     string = "MSpanInuse"
	MSpanSys       string = "MSpanSys"
	Mallocs        string = "Mallocs"
	NextGC         string = "NextGC"
	NumForcedGC    string = "NumForcedGC"
	NumGC          string = "NumGC"
	OtherSys       string = "OtherSys"
	PauseTotalNs   string = "PauseTotalNs"
	StackInuse     string = "StackInuse"
	StackSys       string = "StackSys"
	Sys            string = "Sys"
	TotalAlloc     string = "TotalAlloc"
	PollCount      string = "PollCount"
	RandomValue    string = "RandomValue"
	pollInterval   int    = 2
	reportInterval int    = 10
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

func JSONSendMetrics(url string, metricsData api.Metrics) error {
	data, err := json.Marshal(metricsData)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	request.Header.Set("Content-Type", js)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	defer response.Body.Close()
	return nil
}

func main() {
	var endpoint string
	var nojson bool
	var pInterv int
	var rInterv int
	var err error

	log.SetFlags(log.Ldate | log.Ltime)

	pflag.StringVarP(&endpoint, "endpoint", "a", addressServer, "Used to set the address and port to connect server.")
	pflag.IntVarP(&pInterv, "pollinterval", "p", pollInterval, "User for set poll interval in seconds.")
	pflag.IntVarP(&rInterv, "reportinterval", "r", reportInterval, "User for set report interval (send to srv) in seconds.")
	pflag.BoolVarP(&nojson, "nojson", "n", false, "Use for enable url request")
	pflag.Parse()

	if len(pflag.Args()) != 0 {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	endpointTMP := os.Getenv("ADDRESS")
	if len(endpointTMP) != 0 {
		endpoint = endpointTMP
	}

	pIntervTmp := os.Getenv("POLL_INTERVAL")
	if len(pIntervTmp) != 0 {
		pInterv, err = strconv.Atoi(pIntervTmp)
		if err != nil {
			fmt.Printf("%e\n", err)
			os.Exit(1)
		}
	}

	rIntervTmp := os.Getenv("REPORT_INTERVAL")
	if len(rIntervTmp) != 0 {
		rInterv, err = strconv.Atoi(rIntervTmp)
		if err != nil {
			fmt.Printf("%e\n", err)
			os.Exit(1)
		}
	}

	pollTik := time.NewTicker(time.Duration(pInterv) * time.Second)
	reportTik := time.NewTicker(time.Duration(rInterv) * time.Second)

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
				if nojson {
					err := SendMetrics(fmt.Sprintf("http://%s/update/", endpoint), fmt.Sprintf("gauge/%s/%v", k, v))
					if err != nil {
						log.Print(err)
					}
				} else if !nojson {
					err := JSONSendMetrics(fmt.Sprintf("http://%s/update", endpoint), api.Metrics{MType: "gauge", ID: k, Value: &v})
					if err != nil {
						log.Print(err)
					}
				}
			}
			for k, v := range storage.Counter {
				if nojson {
					err := SendMetrics(fmt.Sprintf("http://%s/update/", endpoint), fmt.Sprintf("counter/%s/%v", k, v))
					if err != nil {
						log.Print(err)
					}
				} else if !nojson {
					err := JSONSendMetrics(fmt.Sprintf("http://%s/update", endpoint), api.Metrics{MType: "counter", ID: k, Delta: &v})
					if err != nil {
						log.Print(err)
					}
				}
			}
		}
	}
}
