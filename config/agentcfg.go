package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"

	"github.com/netzen86/collectmetrics/internal/api"
)

const (
	addressServer  string        = "localhost:8080"
	pollInterval   time.Duration = 2
	reportInterval time.Duration = 10
)

type AgentCfg struct {
	Endpoint        string
	ContentEncoding string
	SignKeyString   string
	PollInterval    int
	Reportinterval  int
}

func GetAgentCfg() (AgentCfg, error) {
	var agentCfg AgentCfg
	var err error

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
	if len(os.Getenv("ADDRESS")) != 0 {
		agentCfg.Endpoint = os.Getenv("ADDRESS")
	}

	if len(os.Getenv("POLL_INTERVAL")) != 0 {
		agentCfg.PollInterval, err = strconv.Atoi(os.Getenv("POLL_INTERVAL"))
		if err != nil {
			return agentCfg, fmt.Errorf("error atoi poll interval %v ", err)
		}
	}

	if len(os.Getenv("REPORT_INTERVAL")) != 0 {
		agentCfg.Reportinterval, err = strconv.Atoi(os.Getenv("REPORT_INTERVAL"))
		if err != nil {
			return agentCfg, fmt.Errorf("error atoi report interval %v ", err)
		}
	}

	if len(os.Getenv("KEY")) != 0 {
		agentCfg.SignKeyString = os.Getenv("KEY")
	}
	return agentCfg, nil
}
