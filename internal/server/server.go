package server

import (
	"context"
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/utils"
)

// создание таблиц counter и gauge
func MakeDBMigrations(ctx context.Context, serverCfg config.ServerCfg,
	srvlog zap.SugaredLogger) error {
	retrybuilder := func() func() error {
		return func() error {
			err := serverCfg.Storage.CreateTables(ctx, srvlog)
			if err != nil {
				log.Println(err)
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
