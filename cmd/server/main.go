// Package main - пакет сервера
// Приложение для получения и храненния метрик.
// Приложение позволяет хранить метрики в текстовом файле, ОЗУ и в базе данных.
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/logger"
	"github.com/netzen86/collectmetrics/internal/repositories/db"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/router"
	"github.com/netzen86/collectmetrics/internal/server"
	"github.com/netzen86/collectmetrics/internal/utils"
)

func main() {
	var cfg config.ServerCfg

	utils.PrintBuildInfos()

	ctx := context.Background()

	// инициализируем логер
	srvlog, err := logger.Logger()
	if err != nil {
		log.Fatalf("error when get logger %v", err)
	}

	// получаем конфиг сервера
	err = cfg.GetServerCfg(srvlog)
	if err != nil {
		srvlog.Fatalf("error when getting config %v ", err)
	}

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

	// сохраняем метрики в файл
	if cfg.StoreInterval != 0 {
		cfg.Wg.Add(1)
		go files.SaveMetrics(cfg.Storage, cfg.FileStoragePathDef,
			cfg.StoreInterval, cfg.ServerCtx, cfg.Wg, srvlog)
	}

	srvlog.Infoln("!!! SERVER START !!!")

	// получаем роутер
	gw := router.GetGateway(cfg, srvlog)
	httpServer := &http.Server{Addr: cfg.Endpoint, Handler: gw}

	// определяем порт для gRPC сервера
	listen, err := net.Listen(config.ProtoTCP, config.EndpointRPC)
	if err != nil {
		srvlog.Fatalf("error when setup net listen %v", err)
	}
	gSRV := server.GetgRPCSrv(cfg)

	go server.GracefulSrv(cfg.Sig, cfg.ServerCtx,
		cfg.ServerStopCtx, httpServer, gSRV, srvlog)

	go func() {
		if err := gSRV.Serve(listen); err != nil {
			srvlog.Fatalf("error when run gRPC server %v", err)
		}
	}()

	// запуск обработчика http запросов
	err = httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		srvlog.Fatalf("error when start server %v ", err)
	}
	<-cfg.ServerCtx.Done()
	cfg.Wg.Wait()
}
