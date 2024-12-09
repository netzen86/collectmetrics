// Package config
// Пакет для конфигурации приложения Агент.
// Получает флаги и переменные окружения.
// Инициализирует некоторые функции.
package config

import (
	"bufio"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/logger"
	"github.com/netzen86/collectmetrics/internal/security"
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
	envPUBKEY          string        = "CRYPTO_KEY"
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

type configAgnFile struct {
	Adderss    string `json:"address,omitempty"`         // аналог переменной окружения ADDRESS или флага -a
	RepInterv  int    `json:"report_interval,omitempty"` // аналог переменной окружения REPORT_INTERVAL или флага -r
	PolIntervv int    `json:"poll_interval,omitempty"`   // аналог переменной окружения POLL_INTERVAL или флага -p
	CryKey     string `json:"crypto_key,omitempty"`      // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
}

// AgentCfg структура для конфигурации Агента
type AgentCfg struct {
	Logger            zap.SugaredLogger `env:"" DefVal:""`
	Endpoint          string            `env:"ADDRESS" DefVal:"localhost:8080"`
	SignKeyString     string            `env:"KEY" DefVal:""`
	ContentEncoding   string            `env:"" DefVal:""`
	PollInterval      int               `env:"POLL_INTERVAL" DefVal:"5"`
	ReportInterval    int               `env:"REPORT_INTERVAL" DefVal:"0"`
	RateLimit         int               `env:"RATE_LIMIT" DefVal:"5"`
	PollTik           time.Duration     `env:"" DefVal:""`
	ReportTik         time.Duration     `env:"" DefVal:""`
	PublicKeyFilename string            `env:"CRYPTO_KEY" DefVal:""`
	PubKey            *rsa.PublicKey    `env:"" DefVal:""`
	AgnFileCfg        string            `env:"" DefVal:""`
}

// функция для получения параметров запуска агента из файла формата json
func getSrvCfgFile(agentCfg *AgentCfg) error {
	var agnCfg configAgnFile
	config, err := os.Open(agentCfg.AgnFileCfg)
	if err != nil {
		return fmt.Errorf("error when read server config file %w", err)
	}
	defer config.Close()

	fileinfo, _ := config.Stat()
	cfgBytes := make([]byte, fileinfo.Size())
	buffer := bufio.NewReader(config)
	_, err = buffer.Read(cfgBytes)
	if err != nil {
		return fmt.Errorf("error when read config file %w", err)
	}
	err = json.Unmarshal(cfgBytes, &agnCfg)
	if err != nil {
		return fmt.Errorf("error when unmarshal config %w", err)
	}

	if agentCfg.Endpoint == addressServerAgent && len(agnCfg.Adderss) != 0 {
		agentCfg.Endpoint = agnCfg.Adderss
	}
	if agentCfg.ReportInterval == int(reportInterval) {
		agentCfg.ReportInterval = agnCfg.RepInterv
	}
	if agentCfg.PollInterval == int(pollInterval) {
		agentCfg.PollInterval = agnCfg.PolIntervv
	}
	if len(agentCfg.PublicKeyFilename) == 0 && len(agnCfg.CryKey) != 0 {
		agentCfg.PublicKeyFilename = agnCfg.CryKey
	}
	return nil
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
	pflag.StringVarP(&agentCfg.ContentEncoding, "contentenc", "e", api.Gz, "Used to set content encoding to connect server.")
	pflag.StringVarP(&agentCfg.SignKeyString, "signkeystring", "k", "", "Used to set key for calc hash.")
	pflag.StringVarP(&agentCfg.PublicKeyFilename, "crypto-key", "s", "", "Load public key for encrypting.")
	pflag.StringVarP(&agentCfg.AgnFileCfg, "config", "c", "", "Load configuration from file.")
	pflag.IntVarP(&agentCfg.PollInterval, "pollinterval", "p", int(pollInterval), "User for set poll interval in seconds.")
	pflag.IntVarP(&agentCfg.ReportInterval, "reportinterval", "r", int(reportInterval), "User for set report interval (send to srv) in seconds.")
	pflag.IntVarP(&agentCfg.RateLimit, "ratelimit", "l", ratelimit, "User for set report interval (send to srv) in seconds.")
	pflag.Parse()

	if len(agentCfg.AgnFileCfg) != 0 {
		err = getSrvCfgFile(&agentCfg)
		if err != nil {
			return AgentCfg{}, fmt.Errorf("when get gonfig from file %w", err)
		}
	}
	log.Println(agentCfg.Endpoint)

	// если переданы аргументы не флаги печатаем подсказку
	if len(pflag.Args()) != 0 {
		pflag.PrintDefaults()
		return AgentCfg{}, fmt.Errorf("accept only dash flags")
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
			return AgentCfg{}, fmt.Errorf("error atoi poll interval %v ", err)
		}
	}

	// получение интервала отправки метрик
	if len(os.Getenv(envRI)) != 0 {
		agentCfg.ReportInterval, err = strconv.Atoi(os.Getenv(envRI))
		if err != nil {
			return AgentCfg{}, fmt.Errorf("error atoi report interval %v ", err)
		}
	}

	// получение лимита отправки метрик
	if len(os.Getenv(envRL)) != 0 {
		agentCfg.RateLimit, err = strconv.Atoi(os.Getenv(envRI))
		if err != nil {
			return AgentCfg{}, fmt.Errorf("error atoi report interval %w ", err)
		}
	}

	// получение ключа для генерации подписи при отправки данных
	if len(os.Getenv(envKey)) != 0 {
		agentCfg.SignKeyString = os.Getenv(envKey)
	}

	// получение публичого ключа для шифрованния
	if len(os.Getenv(envPUBKEY)) != 0 {
		agentCfg.PublicKeyFilename = os.Getenv(envPUBKEY)
	}

	if len(agentCfg.PublicKeyFilename) != 0 {
		agentCfg.PubKey, err = security.ReadPublicKey(agentCfg.PublicKeyFilename)
		if err != nil {
			return AgentCfg{}, fmt.Errorf("error reading public key %w ", err)
		}
	} else {
		agentCfg.PubKey = &rsa.PublicKey{N: big.NewInt(0), E: 0}
	}

	// установка интервалов получения и отправки метрик
	agentCfg.PollTik = time.Duration(agentCfg.PollInterval) * time.Second
	agentCfg.ReportTik = time.Duration(agentCfg.ReportInterval) * time.Second

	return agentCfg, nil
}
