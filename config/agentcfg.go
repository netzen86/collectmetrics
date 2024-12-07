// Package config
// Пакет для конфигурации приложения Агент.
// Получает флаги и переменные окружения.
// Инициализирует некоторые функции.
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

// константы используещиеся для работы Агента
const (
	addressServerAgent string        = "localhost:8080"
	pollInterval       time.Duration = 5
	reportInterval     time.Duration = 0
	ratelimit          int           = 5
	envPI              string        = "POLL_INTERVAL"
	envRI              string        = "REPORT_INTERVAL"
	envRL              string        = "RATE_LIMIT"
	UpdateAddress      string        = "http://%s/update/"
	UpdatesAddress     string        = "http://%s/updates/"
	ProfilerAddr       string        = "localhost:8081"
	Alloc              string        = "Alloc"
	BuckHashSys        string        = "BuckHashSys"
	Frees              string        = "Frees"
	GCCPUFraction      string        = "GCCPUFraction"
	GCSys              string        = "GCSys"
	HeapAlloc          string        = "HeapAlloc"
	HeapIdle           string        = "HeapIdle"
	HeapInuse          string        = "HeapInuse"
	HeapObjects        string        = "HeapObjects"
	HeapReleased       string        = "HeapReleased"
	HeapSys            string        = "HeapSys"
	LastGC             string        = "LastGC"
	Lookups            string        = "Lookups"
	MCacheInuse        string        = "MCacheInuse"
	MCacheSys          string        = "MCacheSys"
	MSpanInuse         string        = "MSpanInuse"
	MSpanSys           string        = "MSpanSys"
	Mallocs            string        = "Mallocs"
	NextGC             string        = "NextGC"
	NumForcedGC        string        = "NumForcedGC"
	NumGC              string        = "NumGC"
	OtherSys           string        = "OtherSys"
	PauseTotalNs       string        = "PauseTotalNs"
	StackInuse         string        = "StackInuse"
	StackSys           string        = "StackSys"
	Sys                string        = "Sys"
	TotalAlloc         string        = "TotalAlloc"
	PollCount          string        = "PollCount"
	RandomValue        string        = "RandomValue"
	TotalMemory        string        = "TotalMemory"
	FreeMemory         string        = "FreeMemory"
	CPUutilization1    string        = "CPUutilization1"
)

// AgentCfg структура для конфигурации Агента
type AgentCfg struct {
	Logger          zap.SugaredLogger `env:"" DefVal:""`
	Endpoint        string            `env:"ADDRESS" DefVal:"localhost:8080"`
	SignKeyString   string            `env:"KEY" DefVal:""`
	ContentEncoding string            `env:"" DefVal:""`
	PollInterval    int               `env:"POLL_INTERVAL" DefVal:"5"`
	ReportInterval  int               `env:"REPORT_INTERVAL" DefVal:"0"`
	RateLimit       int               `env:"RATE_LIMIT" DefVal:"5"`
	PollTik         time.Duration     `env:"" DefVal:""`
	ReportTik       time.Duration     `env:"" DefVal:""`
}

// GetAgentCfg функция получения конфигурации агента.
func GetAgentCfg() (AgentCfg, error) {
	var agentCfg AgentCfg
	var err error

	agentCfg.Logger, err = logger.Logger()
	if err != nil {
		return AgentCfg{}, fmt.Errorf("error when get agent logger %w", err)
	}

	// опредаляем флаги
	pflag.StringVarP(&agentCfg.Endpoint, "endpoint", "a", addressServerAgent, "Used to set the address and port to connect server.")
	pflag.StringVarP(&agentCfg.ContentEncoding, "contentenc", "c", api.Gz, "Used to set content encoding to connect server.")
	pflag.StringVarP(&agentCfg.SignKeyString, "signkeystring", "k", "", "Used to set key for calc hash.")
	pflag.IntVarP(&agentCfg.PollInterval, "pollinterval", "p", int(pollInterval), "User for set poll interval in seconds.")
	pflag.IntVarP(&agentCfg.ReportInterval, "reportinterval", "r", int(reportInterval), "User for set report interval (send to srv) in seconds.")
	pflag.IntVarP(&agentCfg.RateLimit, "ratelimit", "l", ratelimit, "User for set report interval (send to srv) in seconds.")
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
		agentCfg.ReportInterval, err = strconv.Atoi(os.Getenv(envRI))
		if err != nil {
			return agentCfg, fmt.Errorf("error atoi report interval %v ", err)
		}
	}

	// получение лимита отправки метрик
	if len(os.Getenv(envRL)) != 0 {
		agentCfg.RateLimit, err = strconv.Atoi(os.Getenv(envRI))
		if err != nil {
			return agentCfg, fmt.Errorf("error atoi report interval %v ", err)
		}
	}

	// получение ключа для генерации подписи при отправки данных
	if len(os.Getenv(envKey)) != 0 {
		agentCfg.SignKeyString = os.Getenv(envKey)
	}

	// установка интервалов получения и отправки метрик
	agentCfg.PollTik = time.Duration(agentCfg.PollInterval) * time.Second
	agentCfg.ReportTik = time.Duration(agentCfg.ReportInterval) * time.Second

	return agentCfg, nil
}
