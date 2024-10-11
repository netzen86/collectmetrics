package main

import (
	"log"
	"time"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/agent"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/utils"
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

	// устанавливаем для отображения даты и времени в логах
	log.SetFlags(log.Ldate | log.Ltime)

	// установка интервалов получения и отправки метрик
	pollTik := time.NewTicker(time.Duration(agentCfg.PollInterval) * time.Second)
	reportTik := time.NewTicker(time.Duration(agentCfg.Reportinterval) * time.Second)

	for {
		select {
		case <-pollTik.C:
			metrics = agent.CollectMetrics(&counter)
		case <-reportTik.C:
			retrybuilder := func() func() error {
				return func() error {
					err = agent.SendMetrics(metrics, agentCfg.Endpoint, agentCfg.SignKeyString)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				log.Fatal(err)
			}

		}
	}
}
