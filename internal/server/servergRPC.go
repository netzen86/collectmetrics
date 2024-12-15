package server

import (
	"context"
	"log"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/logger"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	pb "github.com/netzen86/collectmetrics/proto/server"
	"google.golang.org/grpc"
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
	response.Metric = &pb.Metrics{}
	_, cntSummed := srv.serverCfg.Storage.(*files.Filestorage)

	srvlog, err := logger.Logger()
	if err != nil {
		log.Fatalf("error when get logger %v", err)
	}

	response.Metric.ID = in.Metric.ID
	response.Metric.MType = in.Metric.MType

	switch {
	case in.Metric.MType == api.Counter:
		err = srv.serverCfg.Storage.UpdateParam(ctx, cntSummed, in.Metric.MType,
			in.Metric.ID, in.Metric.Delta, srvlog)
		if err != nil {
			response.Error = err.Error()
			srvlog.Warnf("error when updating metric %w", err)
		}
		delta, err = srv.serverCfg.Storage.GetCounterMetric(ctx, in.Metric.ID, srvlog)
		response.Metric.Delta = delta
	case in.Metric.MType == api.Gauge:
		err = srv.serverCfg.Storage.UpdateParam(ctx, cntSummed, in.Metric.MType,
			in.Metric.ID, in.Metric.Value, srvlog)
		if err != nil {
			response.Error = err.Error()
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
	response.Metric = &pb.Metrics{}

	srvlog, err := logger.Logger()
	if err != nil {
		log.Fatalf("error when get logger %v", err)
	}

	response.Metric.ID = in.Name
	response.Metric.MType = in.Type

	switch {
	case in.Type == api.Counter:
		delta, err = srv.serverCfg.Storage.GetCounterMetric(ctx, in.Name, srvlog)
		if err != nil {
			response.Error = err.Error()
			srvlog.Warnf("error when getting metiric %w", err)
		}
		srvlog.Infoln("COUNTER VALUE", delta)
		response.Metric.Delta = delta
	case in.Type == api.Gauge:
		value, err = srv.serverCfg.Storage.GetGaugeMetric(ctx, in.Name, srvlog)
		if err != nil {
			response.Error = err.Error()
			srvlog.Warnf("error when getting metiric %w", err)
		}
		srvlog.Infoln("GAUGE VALUE", value)
		response.Metric.Value = value
	}

	return &response, err
}

func (srv *MetricsServer) ListMetricsName(ctx context.Context, in *pb.ListMetricsNameRequest) (*pb.ListMetricsNameResponse, error) {
	var response pb.ListMetricsNameResponse
	var err error
	var metrics api.MetricsMap

	srvlog, err := logger.Logger()
	if err != nil {
		log.Fatalf("error when get logger %v", err)
	}

	metrics, err = srv.serverCfg.Storage.GetAllMetrics(ctx, srvlog)
	if err != nil {
		srvlog.Warnf("error when getting all mtrics name %v", err)
	}

	for key := range metrics.Metrics {
		response.Name = append(response.Name, key)
	}

	return &response, err
}

func GetgRPCSrv(srvCfg config.ServerCfg) *grpc.Server {
	var metricSRV MetricsServer
	metricSRV.serverCfg = &srvCfg
	// создаём gRPC-сервер без зарегистрированной службы
	s := grpc.NewServer()
	// регистрируем сервис
	pb.RegisterMetricServer(s, &metricSRV)
	return s
}
