// Package server - пакет содержит функции для работы сервера
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/utils"
)

// MakeDBMigrations создание таблиц counter и gauge
func MakeDBMigrations(ctx context.Context, serverCfg config.ServerCfg,
	logger zap.SugaredLogger) error {
	retrybuilder := func() func() error {
		return func() error {
			err := serverCfg.Storage.CreateTables(ctx, logger)
			if err != nil {
				logger.Infof("error when creating tables %w", err)
			}
			return err
		}
	}
	err := utils.RetryFunc(retrybuilder)
	if err != nil {
		return fmt.Errorf("tables not created %w", err)
	}
	return nil
}

func RestoreM(ctx context.Context, serverCfg config.ServerCfg,
	srvlog zap.SugaredLogger) error {
	// копируем метрики из файла в хранилище
	if serverCfg.Restore {
		var metrics api.MetricsMap
		metrics.Metrics = make(map[string]api.Metrics)
		srvlog.Infoln("ENTER IN RESTORE")

		err := files.LoadMetric(&metrics, serverCfg.FileStoragePathDef, srvlog)
		if err != nil {
			return fmt.Errorf("error load metrics fom file %w", err)
		}
		for _, metric := range metrics.Metrics {
			if metric.MType == api.Gauge {
				err := serverCfg.Storage.UpdateParam(ctx, false,
					metric.MType, metric.ID, *metric.Value, srvlog)
				if err != nil {
					return fmt.Errorf("error restore lm %s %s : %w", metric.ID, metric.MType, err)
				}
			} else if metric.MType == api.Counter {
				err := serverCfg.Storage.UpdateParam(ctx, false,
					metric.MType, metric.ID, *metric.Delta, srvlog)
				if err != nil {
					return fmt.Errorf("error restore lm %s %s : %w", metric.ID, metric.MType, err)
				}
			}
		}
	}
	return nil
}

func GracefulSrv(sig chan os.Signal, serverCtx context.Context,
	serverStopCtx context.CancelFunc, httpSrv *http.Server, srvlog zap.SugaredLogger) {
	<-sig
	// Shutdown signal with grace period of 30 seconds
	shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
	defer cancel()

	go func() {
		<-shutdownCtx.Done()

		if shutdownCtx.Err() == context.DeadlineExceeded {
			srvlog.Infof("graceful shutdown timed out.. forcing exit.")

		}
	}()

	// Trigger graceful shutdown
	err := httpSrv.Shutdown(shutdownCtx)
	if err != nil {
		srvlog.Infof("error when graceful shutdown %w", err)
	}
	serverStopCtx()
}
