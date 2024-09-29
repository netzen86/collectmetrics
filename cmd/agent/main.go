package main

import (
	"bytes"
	"encoding/hex"
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
	"github.com/netzen86/collectmetrics/internal/security"
	"github.com/netzen86/collectmetrics/internal/utils"

	"github.com/spf13/pflag"
)

const (
	addressServer      string = "localhost:8080"
	templateAddressSrv string = "http://%s/update/"
	updatesAddress     string = "http://%s/updates/"
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

func CollectMetrics(metrics *[]api.Metrics) {
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)

	metFunc := map[string]func() float64{
		Alloc:         func() float64 { return float64(memStats.Alloc) },
		BuckHashSys:   func() float64 { return float64(memStats.BuckHashSys) },
		Frees:         func() float64 { return float64(memStats.Frees) },
		GCCPUFraction: func() float64 { return float64(memStats.GCCPUFraction) },
		GCSys:         func() float64 { return float64(memStats.GCSys) },
		HeapAlloc:     func() float64 { return float64(memStats.HeapAlloc) },
		HeapIdle:      func() float64 { return float64(memStats.HeapIdle) },
		HeapInuse:     func() float64 { return float64(memStats.HeapInuse) },
		HeapObjects:   func() float64 { return float64(memStats.HeapObjects) },
		HeapReleased:  func() float64 { return float64(memStats.HeapReleased) },
		HeapSys:       func() float64 { return float64(memStats.HeapSys) },
		LastGC:        func() float64 { return float64(memStats.LastGC) },
		Lookups:       func() float64 { return float64(memStats.Lookups) },
		MCacheInuse:   func() float64 { return float64(memStats.MCacheInuse) },
		MCacheSys:     func() float64 { return float64(memStats.MCacheSys) },
		MSpanInuse:    func() float64 { return float64(memStats.MSpanInuse) },
		MSpanSys:      func() float64 { return float64(memStats.MSpanSys) },
		Mallocs:       func() float64 { return float64(memStats.Mallocs) },
		NextGC:        func() float64 { return float64(memStats.NextGC) },
		NumForcedGC:   func() float64 { return float64(memStats.NumForcedGC) },
		NumGC:         func() float64 { return float64(memStats.NumGC) },
		OtherSys:      func() float64 { return float64(memStats.OtherSys) },
		PauseTotalNs:  func() float64 { return float64(memStats.PauseTotalNs) },
		StackInuse:    func() float64 { return float64(memStats.StackInuse) },
		StackSys:      func() float64 { return float64(memStats.StackSys) },
		Sys:           func() float64 { return float64(memStats.Sys) },
		TotalAlloc:    func() float64 { return float64(memStats.TotalAlloc) },
		RandomValue:   func() float64 { return rand.Float64() }}

	for k, v := range metFunc {
		value := v()
		*metrics = append(*metrics, api.Metrics{ID: k, MType: gag, Value: &value})
	}
	delta := int64(1)
	*metrics = append(*metrics, api.Metrics{ID: PollCount, MType: cnt, Delta: &delta})
}

func JSONdecode(resp *http.Response) {
	var buf bytes.Buffer
	var metrics []api.Metrics
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
		log.Print("parse json error ", err)
		return
	}
	for _, m := range metrics {
		if m.MType == "counter" {
			log.Printf("%s %v\n", m.ID, *m.Delta)
		}
		if m.MType == "gauge" {
			log.Printf("%s %v\n", m.ID, *m.Value)
		}
	}

}

func JSONSendMetrics(url, signKey string, metrics []api.Metrics) (*http.Response, error) {
	var data, sign []byte
	var err error

	// сериализуем данные в JSON
	data, err = json.Marshal(metrics)
	if err != nil {
		log.Printf("serilazing error: %v\n", err)
		return nil, fmt.Errorf("serilazing error: %v", err)
	}

	// сжимаем данные
	data, err = utils.GzipCompress(data)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	// если передан ключ создаем подпись
	if len(signKey) != 0 {
		sign = security.SignSendData(data, []byte(signKey))
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	request.Header.Add("Content-Encoding", api.Gz)
	request.Header.Add("Content-Type", api.Js)
	request.Header.Add("Accept-Encoding", api.Gz)

	// если передан ключ добавляем подпись к заголовку
	if len(signKey) != 0 {
		request.Header.Add("HashSHA256", hex.EncodeToString(sign))
	}

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

func iterMemStorage(metrics []api.Metrics, endpoint, signKey string) {
	resp, err := JSONSendMetrics(
		fmt.Sprintf(updatesAddress, endpoint),
		signKey, metrics)
	if err != nil {
		log.Println(err)
	}
	JSONdecode(resp)
}

func main() {
	var endpoint string
	var contentEnc string
	var signkeystr string
	var metrics []api.Metrics
	var pInterv int
	var rInterv int
	var err error

	// устанвливаем для отображения даты и времени в логах
	log.SetFlags(log.Ldate | log.Ltime)

	// опредаляем флаги
	pflag.StringVarP(&endpoint, "endpoint", "a", addressServer, "Used to set the address and port to connect server.")
	pflag.StringVarP(&contentEnc, "contentenc", "c", api.Gz, "Used to set content encoding to connect server.")
	pflag.StringVarP(&signkeystr, "signkeystr", "k", "", "Used to set key for calc hash.")
	pflag.IntVarP(&pInterv, "pollinterval", "p", pollInterval, "User for set poll interval in seconds.")
	pflag.IntVarP(&rInterv, "reportinterval", "r", reportInterval, "User for set report interval (send to srv) in seconds.")
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

	signkeystrTmp := os.Getenv("KEY")
	if len(signkeystrTmp) != 0 {
		signkeystr = signkeystrTmp
	}

	pollTik := time.NewTicker(time.Duration(pInterv) * time.Second)
	reportTik := time.NewTicker(time.Duration(rInterv) * time.Second)

	for {
		select {
		case <-pollTik.C:
			CollectMetrics(&metrics)
		case <-reportTik.C:
			iterMemStorage(metrics, endpoint, signkeystr)
		}
	}
}
