package repositories

import (
	"context"

	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

type Repo interface {
	UpdateParam(ctx context.Context, cntSummed bool, metricType, metricName string, metricValue interface{}) error
	GetCounterMetric(ctx context.Context, metricID string) (int64, error)
	GetGaugeMetric(ctx context.Context, metricID string) (float64, error)
	GetStorage(ctx context.Context) (*memstorage.MemStorage, error)
	CreateTables(ctx context.Context) error
}
