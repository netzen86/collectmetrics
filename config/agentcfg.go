// Package config
// Пакет для конфигурации приложения Агент.
// Получает флаги и переменные окружения.
// Инициализирует некоторые функции.
package config

import (
	"bufio"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/logger"
	"github.com/netzen86/collectmetrics/internal/security"
	"github.com/netzen86/collectmetrics/internal/utils"
	pb "github.com/netzen86/collectmetrics/proto/server"
)

// константы используещиеся для работы Агента
const (
	addressServerAgent string        = "localhost:8080"
	AgentgRPCEndpoint  string        = "localhost:3200"
	AgentgRPCProto     string        = "tcp"
	EnablegRPC         bool          = false
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
	Adderss    string `json:"address,omitempty"`
	CryKey     string `json:"crypto_key,omitempty"`
	RepInterv  int    `json:"report_interval,omitempty"`
	PolIntervv int    `json:"poll_interval,omitempty"`
}

// AgentCfg структура для конфигурации Агента
type AgentCfg struct {
	AgentSCtx         context.Context    `env:"" DefVal:""`
	AgentPCtx         context.Context    `env:"" DefVal:""`
	CligRPC           pb.MetricClient    `env:"" DefVal:""`
	Logger            zap.SugaredLogger  `env:"" DefVal:""`
	PubKey            *rsa.PublicKey     `env:"" DefVal:""`
	Sig               chan os.Signal     `env:"" DefVal:""`
	AgentSStopCtx     context.CancelFunc `env:"" DefVal:""`
	AgentPStopCtx     context.CancelFunc `env:"" DefVal:""`
	AgnFileCfg        string             `env:"" DefVal:""`
	ContentEncoding   string             `env:"" DefVal:""`
	PublicKeyFilename string             `env:"CRYPTO_KEY" DefVal:""`
	Endpoint          string             `env:"ADDRESS" DefVal:"localhost:8080"`
	LocalIP           string             `env:"" DefVal:""`
	SignKeyString     string             `env:"KEY" DefVal:""`
	PollInterval      int                `env:"POLL_INTERVAL" DefVal:"5"`
	ReportInterval    int                `env:"REPORT_INTERVAL" DefVal:"0"`
	RateLimit         int                `env:"RATE_LIMIT" DefVal:"5"`
	PollTik           time.Duration      `env:"" DefVal:""`
	ReportTik         time.Duration      `env:"" DefVal:""`
	EnablegRPC        bool               `env:"" DefVal:""`
}

// GetgRPCCli функция для создания клиента gRPC сервера
func GetgRPCCli() (pb.MetricClient, error) {
	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(AgentgRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("error when connect to server %w", err)
	}
	// получаем переменную интерфейсного типа MetricClient,
	// через которую будем отправлять сообщения
	cli := pb.NewMetricClient(conn)
	return cli, nil
}

// функция для получения параметров запуска агента из файла формата json
func getAgnCfgFile(agentCfg *AgentCfg) error {
	var agnCfg configAgnFile
	config, err := os.Open(agentCfg.AgnFileCfg)
	if err != nil {
		return fmt.Errorf("error when read server config file %w", err)
	}
	defer func() {
		err = config.Close()
		if err != nil {
			agentCfg.Logger.Infof("error when close agent cfg %v", err)
		}
	}()

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

func validRateLimit(ratelimit int, logger zap.SugaredLogger) bool {
	if ratelimit == 0 || ratelimit > 32 {
		logger.Infoln("rate limit must be greater than 0 and less than 32")
		return false
	}
	return true
}

func GracefulShutAgent(agentCfg *AgentCfg) {
	agentPCtx, agentPStopCtx := context.WithCancel(context.Background())
	agentCfg.AgentPCtx = agentPCtx
	agentCfg.AgentPStopCtx = agentPStopCtx

	agentSCtx, agentSStopCtx := context.WithCancel(context.Background())
	agentCfg.AgentSCtx = agentSCtx
	agentCfg.AgentSStopCtx = agentSStopCtx

	// Listen for syscall signals for process to interrupt/quit
	agentCfg.Sig = make(chan os.Signal, 1)
	signal.Notify(agentCfg.Sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

// GetAgentCfg функция получения конфигурации агента.
func GetAgentCfg() (AgentCfg, error) {
	var agentCfg AgentCfg
	var err error

	GracefulShutAgent(&agentCfg)

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
	pflag.BoolVarP(&agentCfg.EnablegRPC, "enablegrpc", "g", EnablegRPC, "Use to enable send metiric via gRPC.")
	pflag.Parse()

	if len(agentCfg.AgnFileCfg) != 0 {
		err = getAgnCfgFile(&agentCfg)
		if err != nil {
			return AgentCfg{}, fmt.Errorf("when get gonfig from file %w", err)
		}
	}

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
		agentCfg.PubKey, err = security.ReadPublicKey(agentCfg.PublicKeyFilename, agentCfg.Logger)
		if err != nil {
			return AgentCfg{}, fmt.Errorf("error reading public key %w ", err)
		}
	} else {
		agentCfg.PubKey = &rsa.PublicKey{N: big.NewInt(0), E: 0}
	}

	agentCfg.LocalIP, err = utils.GetLocalIP(agentCfg.Logger)
	if err != nil {
		return AgentCfg{}, fmt.Errorf("error when getting local ip %w ", err)
	}

	if !validRateLimit(agentCfg.RateLimit, agentCfg.Logger) {
		agentCfg.Logger.Infoln("setting rate limit to default value = 5")
		agentCfg.RateLimit = ratelimit
	}

	if agentCfg.EnablegRPC {
		agentCfg.CligRPC, err = GetgRPCCli()
		if err != nil {
			return AgentCfg{}, fmt.Errorf("error when connecting gRPC Server %w ", err)
		}
	}

	// установка интервалов получения и отправки метрик
	agentCfg.PollTik = time.Duration(agentCfg.PollInterval) * time.Second
	agentCfg.ReportTik = time.Duration(agentCfg.ReportInterval) * time.Second

	return agentCfg, nil
}
