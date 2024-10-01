package main

import (
	"log"
	"time"

	agent_config "github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/agent"
	"github.com/netzen86/collectmetrics/internal/api"
)

func main() {
	var metrics []api.Metrics
	var agentCfg agent_config.AgentCfg
	var counter int64
	var err error

	agentCfg, err = agent_config.GetAgentCfg()
	if err != nil {
		log.Fatalf("error on get configuration %v", err)
	}

	// устанвливаем для отображения даты и времени в логах
	log.SetFlags(log.Ldate | log.Ltime)

	pollTik := time.NewTicker(time.Duration(agentCfg.PollInterval) * time.Second)
	reportTik := time.NewTicker(time.Duration(agent.ReportInterval) * time.Second)

	for {
		select {
		case <-pollTik.C:
			agent.CollectMetrics(&metrics, &counter)
		case <-reportTik.C:
			agent.IterMemStorage(metrics, agentCfg.Endpoint, agentCfg.SignKeyString)
		}
	}
}
