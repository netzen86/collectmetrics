package agent

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/security"
	"github.com/netzen86/collectmetrics/internal/utils"
)

const (
	updatesAddress string = "http://%s/updates/"
	gag            string = "gauge"
	cnt            string = "counter"
	alloc          string = "Alloc"
	buckHashSys    string = "BuckHashSys"
	frees          string = "Frees"
	gCCPUFraction  string = "GCCPUFraction"
	gCSys          string = "GCSys"
	heapAlloc      string = "HeapAlloc"
	heapIdle       string = "HeapIdle"
	heapInuse      string = "HeapInuse"
	heapObjects    string = "HeapObjects"
	heapReleased   string = "HeapReleased"
	heapSys        string = "HeapSys"
	lastGC         string = "LastGC"
	lookups        string = "Lookups"
	mCacheInuse    string = "MCacheInuse"
	mCacheSys      string = "MCacheSys"
	mSpanInuse     string = "MSpanInuse"
	mSpanSys       string = "MSpanSys"
	mallocs        string = "Mallocs"
	nextGC         string = "NextGC"
	numForcedGC    string = "NumForcedGC"
	numGC          string = "NumGC"
	otherSys       string = "OtherSys"
	pauseTotalNs   string = "PauseTotalNs"
	stackInuse     string = "StackInuse"
	stackSys       string = "StackSys"
	sys            string = "Sys"
	totalAlloc     string = "TotalAlloc"
	pollCount      string = "PollCount"
	randomValue    string = "RandomValue"
)

// функция сбора метрик
func CollectMetrics(counter *int64) []api.Metrics {
	var memStats runtime.MemStats
	var metrics []api.Metrics
	// runtime.GC()
	runtime.ReadMemStats(&memStats)

	// мапа анонимных функций для сбора метрик
	metFunc := map[string]func() float64{
		alloc:         func() float64 { return float64(memStats.Alloc) },
		buckHashSys:   func() float64 { return float64(memStats.BuckHashSys) },
		frees:         func() float64 { return float64(memStats.Frees) },
		gCCPUFraction: func() float64 { return float64(memStats.GCCPUFraction) },
		gCSys:         func() float64 { return float64(memStats.GCSys) },
		heapAlloc:     func() float64 { return float64(memStats.HeapAlloc) },
		heapIdle:      func() float64 { return float64(memStats.HeapIdle) },
		heapInuse:     func() float64 { return float64(memStats.HeapInuse) },
		heapObjects:   func() float64 { return float64(memStats.HeapObjects) },
		heapReleased:  func() float64 { return float64(memStats.HeapReleased) },
		heapSys:       func() float64 { return float64(memStats.HeapSys) },
		lastGC:        func() float64 { return float64(memStats.LastGC) },
		lookups:       func() float64 { return float64(memStats.Lookups) },
		mCacheInuse:   func() float64 { return float64(memStats.MCacheInuse) },
		mCacheSys:     func() float64 { return float64(memStats.MCacheSys) },
		mSpanInuse:    func() float64 { return float64(memStats.MSpanInuse) },
		mSpanSys:      func() float64 { return float64(memStats.MSpanSys) },
		mallocs:       func() float64 { return float64(memStats.Mallocs) },
		nextGC:        func() float64 { return float64(memStats.NextGC) },
		numForcedGC:   func() float64 { return float64(memStats.NumForcedGC) },
		numGC:         func() float64 { return float64(memStats.NumGC) },
		otherSys:      func() float64 { return float64(memStats.OtherSys) },
		pauseTotalNs:  func() float64 { return float64(memStats.PauseTotalNs) },
		stackInuse:    func() float64 { return float64(memStats.StackInuse) },
		stackSys:      func() float64 { return float64(memStats.StackSys) },
		sys:           func() float64 { return float64(memStats.Sys) },
		totalAlloc:    func() float64 { return float64(memStats.TotalAlloc) },
		randomValue:   func() float64 { return rand.Float64() }}

	for k, v := range metFunc {
		value := v()
		metrics = append(metrics, api.Metrics{ID: k, MType: gag, Value: &value})
	}
	*counter += 1
	metrics = append(metrics, api.Metrics{ID: pollCount, MType: cnt, Delta: counter})
	return metrics
}

// функция для парсинга ответа на запрос обновления метрик
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
	// если данные запакованные распаковываем
	err = utils.SelectDeCoHTTP(&buf, resp)
	if err != nil {
		log.Print("unpack data error", err)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		log.Print("parse json error ", err)
		return
	}
	// типа лог
	for _, m := range metrics {
		if m.MType == "counter" {
			log.Printf("%s %v\n", m.ID, *m.Delta)
		}
		if m.MType == "gauge" {
			log.Printf("%s %v\n", m.ID, *m.Value)
		}
	}

}

// функция для отправки метрик
func JSONSendMetrics(url, signKey string, metrics []api.Metrics) error {
	var data, sign []byte
	var err error

	// сериализуем данные в JSON
	data, err = json.Marshal(metrics)
	if err != nil {
		log.Printf("serilazing error: %v\n", err)
		return fmt.Errorf("serilazing error: %v", err)
	}

	// сжимаем данные
	data, err = utils.GzipCompress(data)
	if err != nil {
		return fmt.Errorf("cannot compress data %v", err)
	}

	// если передан ключ создаем подпись
	if len(signKey) != 0 {
		sign = security.SignSendData(data, []byte(signKey))
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("%v", err)
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
		return fmt.Errorf("%v", err)
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	// defer response.Body.Close()
	JSONdecode(response)
	return nil
}

// функция для отправки метрик
func SendMetrics(metrics []api.Metrics, endpoint, signKey string) error {
	err := JSONSendMetrics(
		fmt.Sprintf(updatesAddress, endpoint),
		signKey, metrics)
	if err != nil {
		return fmt.Errorf("get error when send metric %v", err)
	}
	return nil
}
