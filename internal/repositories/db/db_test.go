// Package db - пакет для работы с хранилищем типа база данных
package db

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/netzen86/collectmetrics/internal/api"
	"go.uber.org/zap"
)

func TestDBStorage_insertData(t *testing.T) {
	type fields struct {
		DB          *sql.DB
		DBconstring string
	}
	type args struct {
		ctx         context.Context
		stmt        string
		metricType  string
		metricName  string
		metricValue interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbstorage := &DBStorage{
				DB:          tt.fields.DB,
				DBconstring: tt.fields.DBconstring,
			}
			if err := dbstorage.insertData(tt.args.ctx, tt.args.stmt, tt.args.metricType, tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.insertData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBStorage_GetAllMetrics(t *testing.T) {
	type fields struct {
		DB          *sql.DB
		DBconstring string
	}
	type args struct {
		ctx    context.Context
		logger zap.SugaredLogger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    api.MetricsMap
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbstorage := &DBStorage{
				DB:          tt.fields.DB,
				DBconstring: tt.fields.DBconstring,
			}
			got, err := dbstorage.GetAllMetrics(tt.args.ctx, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBStorage.GetAllMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBStorage.GetAllMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
