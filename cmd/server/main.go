package main

import (
	"log"
	"net/http"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/router"
)

func main() {
	var cfg config.ServerCfg

	log.Println("!!! SERVER START !!!")

	// получаем конфиг сервера
	err := cfg.GetServerCfg()
	if err != nil {
		log.Fatalf("error when getting config %v ", err)
	}

	// получаем роутер
	gw := router.GetGateway(cfg)

	// восстанавливаем метрики из файла
	if cfg.StoreInterval != 0 {
		go files.SaveMetrics(cfg.Storage, cfg.FileStoragePathDef,
			cfg.Tempfile.Name(), cfg.StorageSelecter, cfg.StoreInterval)
	}

	// запуск обработчика http запросов
	errs := http.ListenAndServe(cfg.Endpoint, gw)
	if errs != nil {
		log.Fatalf("error when start server %v ", errs)
	}
}
