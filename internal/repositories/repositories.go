package repositories

import (
	"context"

	"github.com/netzen86/collectmetrics/internal/api"
)

type Repo interface {
	UpdateParam(ctx context.Context, cntSummed bool, metricType, metricName string, metricValue interface{}) error
	GetCounterMetric(ctx context.Context, metricID string) (int64, error)
	GetGaugeMetric(ctx context.Context, metricID string) (float64, error)
	GetAllMetrics(ctx context.Context) (api.MetricsMap, error)
	CreateTables(ctx context.Context) error
}
