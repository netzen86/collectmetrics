package repositories

import (
	"context"

	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

type Repo interface {
	UpdateParam(ctx context.Context, cntSummed bool, metricType, metricName string, metricValue interface{}) error
	GetMemStorage(ctx context.Context) (*memstorage.MemStorage, error)
}
