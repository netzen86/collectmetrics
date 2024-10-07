package main

import (
	"context"
	"log"
	"net/http"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/router"
)

func main() {
	var cfg config.ServerCfg

	log.Println("!!! SERVER START !!!")

	err := cfg.GetServerCfg()
	if err != nil {
		log.Fatalf("error when getting config %v ", err)
	}

	gw := router.GetGateway(cfg)

	if cfg.StoreInterval != 0 && (cfg.StorageSelecter == "MEMORY" || cfg.StorageSelecter != "DATABASE") {
		storage, err := repositories.Repo.GetStorage(cfg.Storage, context.TODO())
		if err != nil {
			log.Fatalf("error when getting memstorage %v ", err)
		}
		go files.SaveMetrics(storage, cfg.FileStoragePathDef,
			cfg.Tempfile.Name(), cfg.StorageSelecter, cfg.StoreInterval)
	}

	errs := http.ListenAndServe(cfg.Endpoint, gw)
	if errs != nil {
		log.Fatalf("error when start server %v ", errs)
	}
}
