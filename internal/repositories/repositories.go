package repositories

import "context"

type Repo interface {
	UpdateParam(ctx context.Context, metricType, metricName, metricValue string) error
}
