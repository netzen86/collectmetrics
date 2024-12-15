package server

import (
	"context"
	"log"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/logger"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	pb "github.com/netzen86/collectmetrics/proto/server"
)

type MetricsServer struct {
	// нужно встраивать тип pb.Unimplemented
	// для совместимости с будущими версиями
	pb.UnimplementedMetricServer

	// используем для доступа к методар работы с сервером
	serverCfg *config.ServerCfg
}

// AddMetric реализует интерфейс добавления метрик в хранилище.
func (srv *MetricsServer) AddMetric(ctx context.Context, in *pb.AddMetricRequest) (*pb.AddMetircResponse, error) {
	var response pb.AddMetircResponse
	var err error
	var delta int64
	var value float64
	_, cntSummed := srv.serverCfg.Storage.(*files.Filestorage)

	srvlog, err := logger.Logger()
	if err != nil {
		log.Fatalf("error when get logger %v", err)
	}

	response.Metric.ID = in.Metric.ID
	response.Metric.MType = in.Metric.MType
	response.Error = err.Error()

	switch {
	case in.Metric.MType == api.Counter:
		err = srv.serverCfg.Storage.UpdateParam(ctx, cntSummed, in.Metric.MType,
			in.Metric.ID, in.Metric.Delta, srvlog)
		if err != nil {
			srvlog.Warnf("error when updating metric %w", err)
		}
		delta, err = srv.serverCfg.Storage.GetCounterMetric(ctx, in.Metric.ID, srvlog)
		response.Metric.Delta = delta
	case in.Metric.MType == api.Gauge:
		err = srv.serverCfg.Storage.UpdateParam(ctx, cntSummed, in.Metric.MType,
			in.Metric.ID, in.Metric.Value, srvlog)
		if err != nil {
			srvlog.Warnf("error updating metiric %w", err)
		}
		value, err = srv.serverCfg.Storage.GetGaugeMetric(ctx, in.Metric.ID, srvlog)
		response.Metric.Value = value
	}
	return &response, err
}

func (srv *MetricsServer) GetMetric(ctx context.Context, in *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	var response pb.GetMetricResponse
	var err error
	var delta int64
	var value float64

	srvlog, err := logger.Logger()
	if err != nil {
		log.Fatalf("error when get logger %v", err)
	}

	response.Metric.ID = in.Name
	response.Metric.MType = in.Type
	response.Error = err.Error()

	switch {
	case in.Type == api.Counter:
		delta, err = srv.serverCfg.Storage.GetCounterMetric(ctx, in.Name, srvlog)
		if err != nil {
			srvlog.Warnf("error when getting metiric %w", err)
		}
		response.Metric.Delta = delta
	case in.Type == api.Gauge:
		value, err = srv.serverCfg.Storage.GetGaugeMetric(ctx, in.Name, srvlog)
		if err != nil {
			srvlog.Warnf("error when getting metiric %w", err)
		}
		response.Metric.Value = value
	}

	return &response, err
}
