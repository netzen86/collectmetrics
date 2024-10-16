package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/logger"
)

const (
	pollInterval   time.Duration = 2
	reportInterval time.Duration = 10
	envPI          string        = "POLL_INTERVAL"
	envRI          string        = "REPORT_INTERVAL"
	UpdatesAddress string        = "http://%s/updates/"
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
)

type AgentCfg struct {
	Endpoint        string
	ContentEncoding string
	SignKeyString   string
	PollInterval    int
	Reportinterval  int
	PollTik         time.Ticker
	ReportTik       time.Ticker
	Logger          zap.SugaredLogger
}

// функция получения конфигурации сервера
func GetAgentCfg() (AgentCfg, error) {
	var agentCfg AgentCfg
	var err error

	agentCfg.Logger, err = logger.Logger()
	if err != nil {
		return AgentCfg{}, fmt.Errorf("error when get agent logger %w", err)
	}

	// опредаляем флаги
	pflag.StringVarP(&agentCfg.Endpoint, "endpoint", "a", addressServer, "Used to set the address and port to connect server.")
	pflag.StringVarP(&agentCfg.ContentEncoding, "contentenc", "c", api.Gz, "Used to set content encoding to connect server.")
	pflag.StringVarP(&agentCfg.SignKeyString, "signkeystring", "k", "", "Used to set key for calc hash.")
	pflag.IntVarP(&agentCfg.PollInterval, "pollinterval", "p", int(pollInterval), "User for set poll interval in seconds.")
	pflag.IntVarP(&agentCfg.Reportinterval, "reportinterval", "r", int(reportInterval), "User for set report interval (send to srv) in seconds.")
	pflag.Parse()

	// если переданы аргументы не флаги печатаем подсказку
	if len(pflag.Args()) != 0 {
		pflag.PrintDefaults()
		return agentCfg, fmt.Errorf("accept only dash flags")
	}
	// получаем данные для работы програмы из переменных окружения
	// переменные окружения имеют наивысший приоритет
	if len(os.Getenv(envAdd)) != 0 {
		agentCfg.Endpoint = os.Getenv(envAdd)
	}

	// получение интервала сбора метрик
	if len(os.Getenv(envPI)) != 0 {
		agentCfg.PollInterval, err = strconv.Atoi(os.Getenv(envPI))
		if err != nil {
			return agentCfg, fmt.Errorf("error atoi poll interval %v ", err)
		}
	}

	// получение интервала отправки метрик
	if len(os.Getenv(envRI)) != 0 {
		agentCfg.Reportinterval, err = strconv.Atoi(os.Getenv(envRI))
		if err != nil {
			return agentCfg, fmt.Errorf("error atoi report interval %v ", err)
		}
	}
	// получение ключа для генерации подписи при отправки данных
	if len(os.Getenv(envKey)) != 0 {
		agentCfg.SignKeyString = os.Getenv(envKey)
	}

	// установка интервалов получения и отправки метрик
	agentCfg.PollTik = *time.NewTicker(time.Duration(agentCfg.PollInterval) * time.Second)
	agentCfg.ReportTik = *time.NewTicker(time.Duration(agentCfg.Reportinterval) * time.Second)

	return agentCfg, nil
}
