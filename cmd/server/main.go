// Package main - пакет сервера
// Приложение для получения и храненния метрик.
// Приложение позволяет хранить метрики в текстовом файле, ОЗУ и в базе данных.
package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/logger"
	"github.com/netzen86/collectmetrics/internal/repositories/db"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/router"
	"github.com/netzen86/collectmetrics/internal/server"
)

func main() {
	var cfg config.ServerCfg
	ctx := context.Background()

	srvlog, err := logger.Logger()
	if err != nil {
		log.Fatalf("error when get logger %v", err)
	}

	// получаем конфиг сервера
	err = cfg.GetServerCfg(srvlog)
	if err != nil {
		srvlog.Fatalf("error when getting config %v ", err)
	}

	srvlog.Infoln("!!! SERVER START !!!")

	// если хранилище база данных то создаем необходимые таблицы
	_, dbstor := cfg.Storage.(*db.DBStorage)
	if dbstor {
		err = server.MakeDBMigrations(ctx, cfg, srvlog)
		if err != nil {
			srvlog.Fatalf("error when making migration %v ", err)
		}
	}

	// восстанавливаем метрики из файла
	if cfg.Restore {
		err = server.RestoreM(ctx, cfg, srvlog)
		if err != nil {
			srvlog.Fatalf("error when restoring metircs from file %v ", err)
		}
	}

	// получаем роутер
	gw := router.GetGateway(cfg, srvlog)

	// сохраняем метрики в файл
	if cfg.StoreInterval != 0 {
		go files.SaveMetrics(cfg.Storage, cfg.FileStoragePathDef,
			cfg.StorageSelecter, cfg.StoreInterval, srvlog)
	}

	// запуск обработчика http запросов
	err = http.ListenAndServe(cfg.Endpoint, gw)
	if err != nil {
		srvlog.Fatalf("error when start server %v ", err)
	}
}
