package main

import (
	"context"
	"log"
	"net/http"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/repositories/db"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/router"
	"github.com/netzen86/collectmetrics/internal/server"
)

func main() {
	var cfg config.ServerCfg
	ctx := context.Background()

	log.Println("!!! SERVER START !!!")

	// получаем конфиг сервера
	err := cfg.GetServerCfg()
	if err != nil {
		log.Fatalf("error when getting config %v ", err)
	}

	// если хранилище база данных то создаем необходимые таблицы
	_, dbstor := cfg.Storage.(*db.DBStorage)
	if dbstor {
		err = server.MakeDBMigrations(ctx, cfg)
		if err != nil {
			log.Fatalf("error when making migration %v ", err)
		}
	}

	// восстанавливаем метрики из файла
	if cfg.Restore {
		err := server.RestoreM(ctx, cfg)
		if err != nil {
			log.Fatalf("error when restoring metircs from file %v ", err)
		}
	}

	// получаем роутер
	gw := router.GetGateway(cfg)

	// сохраняем метрики из файл
	if cfg.StoreInterval != 0 {
		go files.SaveMetrics(cfg.Storage, cfg.FileStoragePathDef,
			cfg.StorageSelecter, cfg.StoreInterval)
	}

	// запуск обработчика http запросов
	err = http.ListenAndServe(cfg.Endpoint, gw)
	if err != nil {
		log.Fatalf("error when start server %v ", err)
	}
}
