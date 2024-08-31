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
	"github.com/netzen86/collectmetrics/internal/db"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
	"github.com/spf13/pflag"
)

const (
	addressServer      string = "localhost:8080"
	templateAddressSrv string = "http://%s/update/"
	fileSP             string = "metrics.json"
	gag                string = "gauge"
	cnt                string = "counter"
	Alloc              string = "Alloc"
	BuckHashSys        string = "BuckHashSys"
	Frees              string = "Frees"
	GCCPUFraction      string = "GCCPUFraction"
	GCSys              string = "GCSys"
	HeapAlloc          string = "HeapAlloc"
	HeapIdle           string = "HeapIdle"
	HeapInuse          string = "HeapInuse"
	HeapObjects        string = "HeapObjects"
	HeapReleased       string = "HeapReleased"
	HeapSys            string = "HeapSys"
	LastGC             string = "LastGC"
	Lookups            string = "Lookups"
	MCacheInuse        string = "MCacheInuse"
	MCacheSys          string = "MCacheSys"
	MSpanInuse         string = "MSpanInuse"
	MSpanSys           string = "MSpanSys"
	Mallocs            string = "Mallocs"
	NextGC             string = "NextGC"
	NumForcedGC        string = "NumForcedGC"
	NumGC              string = "NumGC"
	OtherSys           string = "OtherSys"
	PauseTotalNs       string = "PauseTotalNs"
	StackInuse         string = "StackInuse"
	StackSys           string = "StackSys"
	Sys                string = "Sys"
	TotalAlloc         string = "TotalAlloc"
	PollCount          string = "PollCount"
	RandomValue        string = "RandomValue"
	pollInterval       int    = 2
	reportInterval     int    = 10
)

func CollectMetrics(storage repositories.Repo, dbconstr string) {
	ctx := context.Background()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)

	storage.UpdateParam(ctx, false, gag, Alloc, float64(memStats.Alloc))
	storage.UpdateParam(ctx, false, gag, BuckHashSys, float64(memStats.BuckHashSys))
	storage.UpdateParam(ctx, false, gag, Frees, float64(memStats.Frees))
	storage.UpdateParam(ctx, false, gag, GCCPUFraction, float64(memStats.GCCPUFraction))
	storage.UpdateParam(ctx, false, gag, GCSys, float64(memStats.GCSys))
	storage.UpdateParam(ctx, false, gag, HeapAlloc, float64(memStats.HeapAlloc))
	storage.UpdateParam(ctx, false, gag, HeapIdle, float64(memStats.HeapIdle))
	storage.UpdateParam(ctx, false, gag, HeapInuse, float64(memStats.HeapInuse))
	storage.UpdateParam(ctx, false, gag, HeapObjects, float64(memStats.HeapObjects))
	storage.UpdateParam(ctx, false, gag, HeapReleased, float64(memStats.HeapReleased))
	storage.UpdateParam(ctx, false, gag, HeapSys, float64(memStats.HeapSys))
	storage.UpdateParam(ctx, false, gag, LastGC, float64(memStats.LastGC))
	storage.UpdateParam(ctx, false, gag, Lookups, float64(memStats.Lookups))
	storage.UpdateParam(ctx, false, gag, MCacheInuse, float64(memStats.MCacheInuse))
	storage.UpdateParam(ctx, false, gag, MCacheSys, float64(memStats.MCacheSys))
	storage.UpdateParam(ctx, false, gag, MSpanInuse, float64(memStats.MSpanInuse))
	storage.UpdateParam(ctx, false, gag, Mallocs, float64(memStats.Mallocs))
	storage.UpdateParam(ctx, false, gag, MSpanSys, float64(memStats.MSpanSys))
	storage.UpdateParam(ctx, false, gag, NextGC, float64(memStats.NextGC))
	storage.UpdateParam(ctx, false, gag, NumForcedGC, float64(memStats.NumForcedGC))
	storage.UpdateParam(ctx, false, gag, NumGC, float64(memStats.NumGC))
	storage.UpdateParam(ctx, false, gag, OtherSys, float64(memStats.OtherSys))
	storage.UpdateParam(ctx, false, gag, PauseTotalNs, float64(memStats.PauseTotalNs))
	storage.UpdateParam(ctx, false, gag, StackInuse, float64(memStats.StackInuse))
	storage.UpdateParam(ctx, false, gag, StackSys, float64(memStats.StackSys))
	storage.UpdateParam(ctx, false, gag, Sys, float64(memStats.Sys))
	storage.UpdateParam(ctx, false, gag, TotalAlloc, float64(memStats.TotalAlloc))
	storage.UpdateParam(ctx, false, gag, RandomValue, rand.Float64())
	storage.UpdateParam(ctx, false, cnt, PollCount, int64(1))

	db.UpdateParamDB(ctx, dbconstr, gag, Alloc, float64(memStats.Alloc))
	db.UpdateParamDB(ctx, dbconstr, gag, BuckHashSys, float64(memStats.BuckHashSys))
	db.UpdateParamDB(ctx, dbconstr, gag, Frees, float64(memStats.Frees))
	db.UpdateParamDB(ctx, dbconstr, gag, GCCPUFraction, float64(memStats.GCCPUFraction))
	db.UpdateParamDB(ctx, dbconstr, gag, GCSys, float64(memStats.GCSys))
	db.UpdateParamDB(ctx, dbconstr, gag, HeapAlloc, float64(memStats.HeapAlloc))
	db.UpdateParamDB(ctx, dbconstr, gag, HeapIdle, float64(memStats.HeapIdle))
	db.UpdateParamDB(ctx, dbconstr, gag, HeapInuse, float64(memStats.HeapInuse))
	db.UpdateParamDB(ctx, dbconstr, gag, HeapObjects, float64(memStats.HeapObjects))
	db.UpdateParamDB(ctx, dbconstr, gag, HeapReleased, float64(memStats.HeapReleased))
	db.UpdateParamDB(ctx, dbconstr, gag, HeapSys, float64(memStats.HeapSys))
	db.UpdateParamDB(ctx, dbconstr, gag, LastGC, float64(memStats.LastGC))
	db.UpdateParamDB(ctx, dbconstr, gag, Lookups, float64(memStats.Lookups))
	db.UpdateParamDB(ctx, dbconstr, gag, MCacheInuse, float64(memStats.MCacheInuse))
	db.UpdateParamDB(ctx, dbconstr, gag, MCacheSys, float64(memStats.MCacheSys))
	db.UpdateParamDB(ctx, dbconstr, gag, MSpanInuse, float64(memStats.MSpanInuse))
	db.UpdateParamDB(ctx, dbconstr, gag, Mallocs, float64(memStats.Mallocs))
	db.UpdateParamDB(ctx, dbconstr, gag, MSpanSys, float64(memStats.MSpanSys))
	db.UpdateParamDB(ctx, dbconstr, gag, NextGC, float64(memStats.NextGC))
	db.UpdateParamDB(ctx, dbconstr, gag, NumForcedGC, float64(memStats.NumForcedGC))
	db.UpdateParamDB(ctx, dbconstr, gag, NumGC, float64(memStats.NumGC))
	db.UpdateParamDB(ctx, dbconstr, gag, OtherSys, float64(memStats.OtherSys))
	db.UpdateParamDB(ctx, dbconstr, gag, PauseTotalNs, float64(memStats.PauseTotalNs))
	db.UpdateParamDB(ctx, dbconstr, gag, StackInuse, float64(memStats.StackInuse))
	db.UpdateParamDB(ctx, dbconstr, gag, StackSys, float64(memStats.StackSys))
	db.UpdateParamDB(ctx, dbconstr, gag, Sys, float64(memStats.Sys))
	db.UpdateParamDB(ctx, dbconstr, gag, TotalAlloc, float64(memStats.TotalAlloc))
	db.UpdateParamDB(ctx, dbconstr, gag, RandomValue, rand.Float64())
	db.UpdateParamDB(ctx, dbconstr, cnt, PollCount, int64(1))
}

func SendMetrics(url, metricData string) error {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", url, metricData), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", api.Th)
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

func GetAccEnc(url, contEnc string) (string, error) {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	request.Header.Add("Content-Encoding", contEnc)
	request.Header.Add("Content-Type", api.Js)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	defer response.Body.Close()

	log.Printf("%s %s\n", "GetAccEnc", response.Status)
	encoding := response.Header.Get("Accept-Encoding")
	return encoding, nil
}

func JSONdecode(resp *http.Response) {
	var buf bytes.Buffer
	var metrics api.Metrics
	if resp == nil {
		log.Print("error nil response")
		return
	}
	defer resp.Body.Close()
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		log.Print("reading body error ", err)
		return
	}
	// если данные запакованные
	err = utils.SelectDeCoHTTP(&buf, resp)
	if err != nil {
		log.Print("unpack data error", err)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		log.Print("parse json error", err)
		return
	}
	if metrics.MType == "counter" {
		log.Printf("%s %v\n", metrics.ID, *metrics.Delta)
	}
	if metrics.MType == "gauge" {
		log.Printf("%s %v\n", metrics.ID, *metrics.Value)
	}
}

func JSONSendMetrics(url, ce string, metricsData api.Metrics) (*http.Response, error) {
	// if metricsData.MType == "counter" {
	// 	log.Println("JSON SEND ", metricsData.MType, *metricsData.Delta)
	// }
	// получаем от сервера ответ о поддерживаемыж методах сжатия
	encoding, err := GetAccEnc(url, ce)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	// сериализуем данные в JSON
	data, err := json.Marshal(metricsData)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	metricsData.Clean()
	// если сервер поддерживает сжатие сжимаем данные
	if encoding == "gzip" {
		data, err = utils.GzipCompress(data)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	if encoding == "gzip" {
		request.Header.Add("Content-Encoding", api.Gz)
	}

	request.Header.Add("Content-Type", api.Js)
	request.Header.Add("Accept-Encoding", api.Gz)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	if response.StatusCode != 200 {
		return nil, errors.New(response.Status)
	}
	// defer response.Body.Close()
	return response, nil
}

func main() {
	var endpoint string
	var contentEnc string
	var dbconstring string
	var fileStoragePath string
	var nojson bool
	var pInterv int
	var rInterv int
	var err error
	ctx := context.TODO()
	// устанвливаем для отображения даты и времени в логах
	log.SetFlags(log.Ldate | log.Ltime)

	// опредаляем флаги
	pflag.StringVarP(&endpoint, "endpoint", "a", addressServer, "Used to set the address and port to connect server.")
	pflag.StringVarP(&contentEnc, "contentenc", "c", api.Gz, "Used to set content encoding to connect server.")
	pflag.StringVarP(&fileStoragePath, "filepath", "f", fileSP, "Used to set file path to save metrics.")
	pflag.StringVarP(&dbconstring, "dbconstring", "d", db.DataBaseConString, "Used to set file path to save metrics.")
	pflag.IntVarP(&pInterv, "pollinterval", "p", pollInterval, "User for set poll interval in seconds.")
	pflag.IntVarP(&rInterv, "reportinterval", "r", reportInterval, "User for set report interval (send to srv) in seconds.")
	pflag.BoolVarP(&nojson, "nojson", "n", false, "Use for enable url request")
	pflag.Parse()

	// если переданы аргументы не флаги печатаем подсказку
	if len(pflag.Args()) != 0 {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// получаем данные для работы програмы из переменных окружения
	// переменные окружения имеют наивысший приоритет
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

	fileStoragePathTMP := os.Getenv("FILE_STORAGE_PATH")
	if len(fileStoragePathTMP) != 0 {
		fileStoragePath = fileStoragePathTMP
	}

	dbaddressTMP := os.Getenv("DATABASE_DSN")
	if len(dbaddressTMP) != 0 {
		dbconstring = dbaddressTMP
	}

	pollTik := time.NewTicker(time.Duration(pInterv) * time.Second)
	reportTik := time.NewTicker(time.Duration(rInterv) * time.Second)

	err = db.CreateTables(ctx, dbconstring)
	if err != nil {
		log.Fatal(err)
	}

	storage, err := memstorage.NewMemStorage()
	if err != nil {
		panic("couldn't alloc mem")
	}

	for {
		select {
		case <-pollTik.C:
			CollectMetrics(storage, db.DataBaseConString)
		case <-reportTik.C:
			for k, v := range storage.Gauge {
				if nojson {
					err := SendMetrics(fmt.Sprintf(templateAddressSrv, endpoint), fmt.Sprintf("gauge/%s/%v", k, v))
					if err != nil {
						log.Println(err)
					}
				} else if !nojson {
					resp, err := JSONSendMetrics(
						fmt.Sprintf(templateAddressSrv, endpoint),
						contentEnc,
						api.Metrics{MType: "gauge", ID: k, Value: &v})
					if err != nil {
						log.Println("Gauge error ", err)
					}
					JSONdecode(resp)
				}
			}
			for k, v := range storage.Counter {
				if nojson {
					err := SendMetrics(fmt.Sprintf(templateAddressSrv, endpoint), fmt.Sprintf("counter/%s/%v", k, v))
					if err != nil {
						log.Println(err)
					}
				} else if !nojson {
					log.Println("counter value !!!", v)
					resp, err := JSONSendMetrics(
						fmt.Sprintf(templateAddressSrv, endpoint),
						contentEnc,
						api.Metrics{MType: "counter", ID: k, Delta: &v})
					if err != nil {
						log.Println(err)
					}
					JSONdecode(resp)
				}
			}
		}
	}
}
