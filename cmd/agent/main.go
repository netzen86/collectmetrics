// Приложение для сбора и отправки меткрик
package main

import (
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/agent"
	"github.com/netzen86/collectmetrics/internal/logger"
)

func main() {
	var agentCfg config.AgentCfg
	var err error

	agnlog, err := logger.Logger()
	if err != nil {
		log.Fatalf("error when get logger %v", err)
	}

	// получаем конфиг агента
	agentCfg, err = config.GetAgentCfg()
	if err != nil {
		agnlog.Fatalf("error on get configuration %v", err)
	}

	// запускаем агента
	err = agent.RunAgent(agentCfg)
	if err != nil {
		agnlog.Fatalf("agent don't send metrics %v", err)
	}

	agnlog.Infoln("RUNNIG SRV FOR PROFILING")
	err = http.ListenAndServe(config.ProfilerAddr, nil)
	if err != nil {
		agnlog.Fatalf("error when start server %v ", err)
	}
}
