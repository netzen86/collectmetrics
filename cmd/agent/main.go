package main

import (
	"log"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/agent"
	"github.com/netzen86/collectmetrics/internal/api"
)

func main() {
	var metrics []api.Metrics
	var agentCfg config.AgentCfg
	var counter int64
	var err error

	// получаем конфиг агента
	agentCfg, err = config.GetAgentCfg()
	if err != nil {
		log.Fatalf("error on get configuration %v", err)
	}

	// запускаем агента
	err = agent.RunAgent(metrics, agentCfg, &counter)
	if err != nil {
		agentCfg.Logger.Fatalf("agent don't send metrics %v", err)
	}
}
