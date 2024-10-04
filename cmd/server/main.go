package main

import (
	"log"
	"net/http"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/router"
)

func main() {
	log.Println("!!! SERVER START !!!")

	cfg, err := config.GetServerCfg()
	if err != nil {
		log.Fatalf("error when getting config %v ", err)
	}

	gw := router.GetGateway(cfg)

	if cfg.StoreInterval != 0 {
		go files.SaveMetrics(cfg.MemStorage, cfg.FileStoragePathDef,
			cfg.Tempfile.Name(), cfg.StorageSelecter, cfg.StoreInterval)
	}

	errs := http.ListenAndServe(cfg.Endpoint, gw)
	if errs != nil {
		log.Fatalf("error when start server %v ", errs)
	}
}
