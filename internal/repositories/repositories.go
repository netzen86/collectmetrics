package repositories

import (
	"context"

	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
)

type Repo interface {
	UpdateParam(ctx context.Context, cntSummed bool, metricType, metricName string, metricValue interface{}, srvlog zap.SugaredLogger) error
	GetCounterMetric(ctx context.Context, metricID string, srvlog zap.SugaredLogger) (int64, error)
	GetGaugeMetric(ctx context.Context, metricID string, srvlog zap.SugaredLogger) (float64, error)
	GetAllMetrics(ctx context.Context, srvlog zap.SugaredLogger) (api.MetricsMap, error)
	CreateTables(ctx context.Context, srvlog zap.SugaredLogger) error
}
