package memstorage

import (
	"context"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		want    *MemStorage
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMemStorage()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_UpdateParam(t *testing.T) {
	type fields struct {
		Gauge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		ctx         context.Context
		metricValue interface{}
		srvlog      zap.SugaredLogger
		metricType  string
		metricName  string
	}
	tests := []struct {
		args    args
		fields  fields
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MemStorage{
				Gauge:   tt.fields.Gauge,
				Counter: tt.fields.Counter,
			}
			if err := storage.UpdateParam(tt.args.ctx, false,
				tt.args.metricType, tt.args.metricName, tt.args.metricValue, tt.args.srvlog); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.UpdateParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
