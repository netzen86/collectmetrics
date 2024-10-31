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
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/security"
	"github.com/netzen86/collectmetrics/internal/utils"
)

type gaugeJobs struct {
	mName    string
	function func() float64
}

type counterJobs struct {
	mName    string
	counter  *int64
	function func(counter *int64) int64
}

func workerGauge(job <-chan gaugeJobs, results chan<- api.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	memStat, ok := <-job
	if !ok {
		return
	}
	value := memStat.function()
	results <- api.Metrics{ID: memStat.mName, MType: api.Gauge, Value: &value}
}

func workerCounter(job <-chan counterJobs, results chan<- api.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	cntData, ok := <-job
	if !ok {
		return
	}
	cnt := int64(0)
	cntData.counter = &cnt
	delta := cntData.function(cntData.counter)
	results <- api.Metrics{ID: cntData.mName, MType: api.Counter, Delta: &delta}
}

// функция сбора метрик
func CollectMetrics(counter *int64, numJobs int, results chan api.Metrics) error {
	log.Println("RUN COLLECT METRICS")

	var memStats runtime.MemStats

	mem, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("error when getting ext mem stat %w", err)
	}

	cpuStat, err := cpu.Counts(true)
	if err != nil {
		return fmt.Errorf("error when getting cpu stat %w", err)
	}

	wg := &sync.WaitGroup{}
	jobsGauge := make(chan gaugeJobs, numJobs)
	jobsCounter := make(chan counterJobs, numJobs)

	for w := 1; w <= numJobs; w++ {
		wg.Add(1)
		go workerGauge(jobsGauge, results, wg)
	}

	for w := 1; w <= 1; w++ {
		wg.Add(1)
		go workerCounter(jobsCounter, results, wg)
	}

	runtime.ReadMemStats(&memStats)

	// мапа анонимных функций для сбора метрик
	gaugeFunc := map[string]func() float64{
		config.Alloc:           func() float64 { return float64(memStats.Alloc) },
		config.BuckHashSys:     func() float64 { return float64(memStats.BuckHashSys) },
		config.Frees:           func() float64 { return float64(memStats.Frees) },
		config.GCCPUFraction:   func() float64 { return float64(memStats.GCCPUFraction) },
		config.GCSys:           func() float64 { return float64(memStats.GCSys) },
		config.HeapAlloc:       func() float64 { return float64(memStats.HeapAlloc) },
		config.HeapIdle:        func() float64 { return float64(memStats.HeapIdle) },
		config.HeapInuse:       func() float64 { return float64(memStats.HeapInuse) },
		config.HeapObjects:     func() float64 { return float64(memStats.HeapObjects) },
		config.HeapReleased:    func() float64 { return float64(memStats.HeapReleased) },
		config.HeapSys:         func() float64 { return float64(memStats.HeapSys) },
		config.LastGC:          func() float64 { return float64(memStats.LastGC) },
		config.Lookups:         func() float64 { return float64(memStats.Lookups) },
		config.MCacheInuse:     func() float64 { return float64(memStats.MCacheInuse) },
		config.MCacheSys:       func() float64 { return float64(memStats.MCacheSys) },
		config.MSpanInuse:      func() float64 { return float64(memStats.MSpanInuse) },
		config.MSpanSys:        func() float64 { return float64(memStats.MSpanSys) },
		config.Mallocs:         func() float64 { return float64(memStats.Mallocs) },
		config.NextGC:          func() float64 { return float64(memStats.NextGC) },
		config.NumForcedGC:     func() float64 { return float64(memStats.NumForcedGC) },
		config.NumGC:           func() float64 { return float64(memStats.NumGC) },
		config.OtherSys:        func() float64 { return float64(memStats.OtherSys) },
		config.PauseTotalNs:    func() float64 { return float64(memStats.PauseTotalNs) },
		config.StackInuse:      func() float64 { return float64(memStats.StackInuse) },
		config.StackSys:        func() float64 { return float64(memStats.StackSys) },
		config.Sys:             func() float64 { return float64(memStats.Sys) },
		config.TotalAlloc:      func() float64 { return float64(memStats.TotalAlloc) },
		config.RandomValue:     func() float64 { return rand.Float64() },
		config.TotalMemory:     func() float64 { return float64(mem.Total) },
		config.FreeMemory:      func() float64 { return float64(mem.Free) },
		config.CPUutilization1: func() float64 { return float64(cpuStat) }}

	// мапа анонимных функций для сбора метрик
	counterFunc := map[string]func(coutner *int64) int64{
		config.PollCount: func(counter *int64) int64 { *counter += 1; return *counter },
	}
	// в канал задач отправляем задачи
	for k, v := range gaugeFunc {
		jobsGauge <- gaugeJobs{mName: k, function: v}
	}
	close(jobsGauge)

	for k, v := range counterFunc {
		jobsCounter <- counterJobs{mName: k, counter: counter, function: v}
	}
	close(jobsCounter)

	wg.Wait()
	close(results)
	return nil
}

// функция для парсинга ответа на запрос обновления метрик
func JSONdecode(resp *http.Response, logger zap.SugaredLogger) {
	var buf bytes.Buffer
	var metrics []api.Metrics
	if resp == nil {
		logger.Infoln("error nil response")
		return
	}
	defer resp.Body.Close()
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		logger.Infoln("reading body error ", err)
		return
	}
	// если данные запакованные распаковываем
	err = utils.SelectDeCoHTTP(&buf, resp, logger)
	if err != nil {
		logger.Infoln("unpack data error", err)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		logger.Infoln("parse json error ", err)
		return
	}

	// типа лог
	for _, m := range metrics {
		if m.MType == api.Counter {
			logger.Infof("%s %v", m.ID, *m.Delta)
		}
		if m.MType == api.Gauge {
			logger.Infof("%s %v", m.ID, *m.Value)
		}
	}

}

// функция для отправки метрик
func JSONSendMetrics(url, signKey string, metricsChan chan api.Metrics, logger zap.SugaredLogger) error {
	var metrics []api.Metrics
	var data, sign []byte
	var err error

	for metric := range metricsChan {
		metrics = append(metrics, metric)
	}

	// сериализуем данные в JSON
	data, err = json.Marshal(metrics)
	if err != nil {
		logger.Infof("serilazing error: %v\n", err)
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

	// создаем реквест
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// добавляем данные в заголовок запроса
	request.Header.Add("Content-Encoding", api.Gz)
	request.Header.Add("Content-Type", api.Js)
	request.Header.Add("Accept-Encoding", api.Gz)

	// если передан ключ добавляем подпись к заголовку
	if len(signKey) != 0 {
		request.Header.Add("HashSHA256", hex.EncodeToString(sign))
	}

	// создаем http клиент
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	// defer response.Body.Close()
	JSONdecode(response, logger)
	return nil
}

// функция для отправки метрик
func SendMetrics(metrics chan api.Metrics, endpoint,
	signKey string, logger zap.SugaredLogger,
	errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	err := JSONSendMetrics(
		fmt.Sprintf(config.UpdatesAddress, endpoint),
		signKey, metrics, logger)
	if err != nil {
		errCh <- fmt.Errorf("get error when send metric %v", err)
		return
	}
}

func RunAgent(agentCfg config.AgentCfg) error {
	var metrics chan api.Metrics
	var errCh chan error
	counter := int64(0)
	numJobs := 32

	for {
		select {
		case <-agentCfg.PollTik.C:
			metrics = make(chan api.Metrics, numJobs)
			CollectMetrics(&counter, numJobs, metrics)
		case <-agentCfg.ReportTik.C:
			errCh = make(chan error)
			wg := &sync.WaitGroup{}

			retrybuilder := func() func() error {
				return func() error {
					wg.Add(1)
					go SendMetrics(metrics, agentCfg.Endpoint,
						agentCfg.SignKeyString, agentCfg.Logger, errCh, wg)
					if len(errCh) > 0 {
						agentCfg.Logger.Infof("error when sm in internal/agent %w", <-errCh)
					}
					wg.Wait()
					return nil
				}
			}
			err := utils.RetryFunc(retrybuilder)
			if err != nil {
				return fmt.Errorf("fail when sm in agent %w", err)
			}
		}
	}
}
