// Package agent - пакет содержит функции для работы агента
package agent

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/security"
	"github.com/netzen86/collectmetrics/internal/utils"
	pb "github.com/netzen86/collectmetrics/proto/server"
)

type gaugeJobs struct {
	function func() float64
	mName    string
}

type counterJobs struct {
	counter  *int64
	function func(counter *int64) int64
	mName    string
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
	// cnt := int64(0)
	// cntData.counter = &cnt
	delta := cntData.function(cntData.counter)
	*cntData.counter = delta
	results <- api.Metrics{ID: cntData.mName, MType: api.Counter, Delta: &delta}
}

// CollectMetrics функция сбора метрик
func CollectMetrics(counter *int64, agentCfg config.AgentCfg,
	results chan api.Metrics, errCh chan<- error, rwg *sync.WaitGroup) {
	defer rwg.Done()
	var memStats runtime.MemStats
	shutdown := false

	if results == nil {
		errCh <- fmt.Errorf("channel closed")
		return
	}

	mem, err := mem.VirtualMemory()
	if err != nil {
		errCh <- fmt.Errorf("error when getting ext mem stat %w", err)
		close(results)
	}

	cpuStat, err := cpu.Counts(true)
	if err != nil {
		errCh <- fmt.Errorf("error when getting cpu stat %w", err)
		close(results)
	}

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

	for !shutdown {
		<-time.After(agentCfg.PollTik)
		agentCfg.Logger.Infoln("COLLECTING METRIC")

		runtime.ReadMemStats(&memStats)

		wg := &sync.WaitGroup{}
		jobsGauge := make(chan gaugeJobs, len(gaugeFunc)+1)
		jobsCounter := make(chan counterJobs, len(counterFunc)+1)

		for range len(gaugeFunc) + 1 {
			wg.Add(1)
			go workerGauge(jobsGauge, results, wg)
		}

		for range len(counterFunc) + 1 {
			wg.Add(1)
			go workerCounter(jobsCounter, results, wg)
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
		select {
		case <-agentCfg.AgentPCtx.Done():
			shutdown = true
			close(results)
			agentCfg.Logger.Info("-=*** STOP POOLING METRICS ***=-")
			stopWithTimer(agentCfg.AgentSCtx, agentCfg.AgentSStopCtx, agentCfg.Logger)
		default:
		}
	}
}

// JSONdecode функция для парсинга ответа на запрос обновления метрик
func JSONdecode(resp *http.Response, logger zap.SugaredLogger) {
	var buf bytes.Buffer
	var metrics api.Metrics
	var err error

	if resp == nil {
		logger.Infoln("error nil response")
		return
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			logger.Errorf("error when closing body %v", err)
		}
	}()

	_, err = buf.ReadFrom(resp.Body)
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

	if metrics.MType == api.Counter {
		logger.Infof("%s %v", metrics.ID, *metrics.Delta)
	}
	if metrics.MType == api.Gauge {
		logger.Infof("%s %v", metrics.ID, *metrics.Value)
	}
}

// JSONSendMetrics функция для отправки метрик
func JSONSendMetrics(url, signKey, localIP string, metrics api.Metrics, pubKey *rsa.PublicKey, logger zap.SugaredLogger) error {
	var data, sign []byte
	var err error

	// сериализуем данные в JSON
	data, err = json.Marshal(metrics)
	if err != nil {
		logger.Infof("serilazing error: %v\n", err)
		return fmt.Errorf("serilazing error: %v", err)
	}

	// если перадан публичнный ключ - шифруем контент
	if pubKey.Size() != 0 {
		data, err = security.EncryptMetic(data, pubKey)
		if err != nil {
			return fmt.Errorf("cannot encrypt metric %w", err)
		}
	}

	// сжимаем данные
	data, err = utils.GzipCompress(data)
	if err != nil {
		return fmt.Errorf("cannot compress metirc %w", err)
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
	request.Header.Add(api.ACLHeader, localIP)
	// если передан публичный ключ добавляем к заголовку парамер что контент зашифрован
	if pubKey.Size() != 0 {
		request.Header.Add("CryptRSA", api.CryptRSA)
	}

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
	defer func() {
		err = response.Body.Close()
		if err != nil {
			logger.Infof("error when body closing %v", err)
		}
	}()

	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	JSONdecode(response, logger)
	return nil
}

func workerSM(jobs <-chan api.Metrics, endpoint, signKey, localIP string,
	pubKey *rsa.PublicKey, logger zap.SugaredLogger, gRPCCli pb.MetricClient, enablegRPC bool,
	errCh chan<- error, wg *sync.WaitGroup) {
	var err error
	ctx := context.Background()
	defer wg.Done()
	metric, ok := <-jobs
	if !ok {
		return
	}

	retrybuilder := func() func() error {
		return func() error {
			switch {
			case enablegRPC:
				var pbMetric pb.AddMetricRequest
				var response *pb.AddMetircResponse
				pbMetric.Metric = &pb.Metrics{}

				pbMetric.Metric.Id = metric.ID
				pbMetric.Metric.Mtype = metric.MType

				if metric.MType == api.Counter {
					pbMetric.Metric.Delta = *metric.Delta
				} else if metric.MType == api.Gauge {
					pbMetric.Metric.Value = *metric.Value
				}

				response, err = gRPCCli.AddMetric(ctx, &pbMetric)
				if err != nil {
					logger.Infof("error when sm gRPC in internal/agent %v", err)
				}
				logger.Infoln(response.Metric.Id, response.Metric.Mtype,
					response.Metric.Delta, response.Metric.Value)
			default:
				err = JSONSendMetrics(
					fmt.Sprintf(config.UpdateAddress, endpoint),
					signKey, localIP, metric, pubKey, logger)

				if err != nil {
					logger.Infof("error when sm in internal/agent %v", err)
				}
			}
			return nil
		}
	}
	err = utils.RetryFunc(retrybuilder)
	if err != nil {
		errCh <- fmt.Errorf("fail when sm in agent %w", err)
		return
	}
}

// SendMetrics функция для отправки метрик
func SendMetrics(metrics <-chan api.Metrics, agentCfg config.AgentCfg,
	errCh chan<- error, rwg *sync.WaitGroup) {
	defer rwg.Done()
	jobs := make(chan api.Metrics, agentCfg.RateLimit)
	wg := sync.WaitGroup{}
	shutdown := false

	for !shutdown {

		<-time.After(agentCfg.ReportTik)

		for range agentCfg.RateLimit {
			wg.Add(1)
			go workerSM(jobs, agentCfg.Endpoint, agentCfg.SignKeyString, agentCfg.LocalIP,
				agentCfg.PubKey, agentCfg.Logger, agentCfg.CligRPC,
				agentCfg.EnablegRPC, errCh, &wg)
		}

		for range agentCfg.RateLimit {
			metric, ok := <-metrics
			if !ok {
				break
			}
			jobs <- metric
		}
		wg.Wait()

		select {
		case <-agentCfg.AgentSCtx.Done():
			agentCfg.Logger.Info("-=*** STOP SENDING METIRICS ***=-")
			close(jobs)
			shutdown = true
		default:
		}
	}
}

func sigMon(sig chan os.Signal, agentCtx context.Context,
	agentStopCtx context.CancelFunc, logger zap.SugaredLogger) {
	<-sig
	stopWithTimer(agentCtx, agentStopCtx, logger)
}

func stopWithTimer(agentCtx context.Context,
	agentStopCtx context.CancelFunc, logger zap.SugaredLogger) {

	// Shutdown signal with grace period of 30 seconds
	shutdownCtx, cancel := context.WithTimeout(agentCtx, 30*time.Second)
	defer cancel()

	go func() {
		<-shutdownCtx.Done()

		if shutdownCtx.Err() == context.DeadlineExceeded {
			logger.Infof("graceful shutdown timed out.. forcing exit.")

		}
	}()
	agentStopCtx()
}

func RunAgent(agentCfg config.AgentCfg) error {
	counter := int64(0)
	numJobs := 32
	errCh := make(chan error)
	metrics := make(chan api.Metrics, numJobs)
	rwg := &sync.WaitGroup{}

	go sigMon(agentCfg.Sig, agentCfg.AgentPCtx, agentCfg.AgentPStopCtx, agentCfg.Logger)

	rwg.Add(1)
	go CollectMetrics(&counter, agentCfg, metrics, errCh, rwg)

	rwg.Add(1)
	go SendMetrics(metrics, agentCfg, errCh, rwg)

	rwg.Wait()
	return nil
}
