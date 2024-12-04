// Package db - пакет для работы с хранилищем типа база данных
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/utils"
)

// DBStorage адрес для подключения к базе данных -  "postgres://postgres:collectmetrics@localhost/collectmetrics?sslmode=disable"
type DBStorage struct {
	DB          *sql.DB
	DBconstring string
}

// NewDBStorage функция подключения к базе данных, param = строка для подключения к БД
func NewDBStorage(ctx context.Context, param string) (*DBStorage, error) {
	var dbstorage DBStorage
	var err error
	dbstorage.DB, err = sql.Open("pgx", param)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to data base %w", err)
	}
	err = dbstorage.DB.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot ping data base %w", err)
	}
	return &dbstorage, nil
}

// функция для вставки данных в базу данных
func (dbstorage *DBStorage) insertData(ctx context.Context, stmt, metricType,
	metricName string, metricValue interface{}) error {
	if metricType == api.Gauge {

		value, err := utils.ParseValGag(metricValue)
		if err != nil {
			return err
		}
		_, err = dbstorage.DB.ExecContext(ctx, stmt, metricName, value)
		if err != nil {
			return fmt.Errorf("insert in table error - %w", err)
		}
	}

	if metricType == api.Counter {
		delta, err := utils.ParseValCnt(metricValue)
		if err != nil {
			return err
		}
		_, err = dbstorage.DB.ExecContext(ctx, stmt, metricName, delta)
		if err != nil {
			return fmt.Errorf("insert in table error - %w", err)
		}

	}
	return nil
}

func (dbstorage *DBStorage) GetAllMetrics(ctx context.Context, logger zap.SugaredLogger) (api.MetricsMap, error) {
	var metrics api.MetricsMap
	metrics.Metrics = make(map[string]api.Metrics)

	smtp := `
	SELECT name, value, 'gauge' as type
	FROM gauge
	UNION all
	SELECT name, delta, 'counter' as type
	FROM counter;`

	rows, err := dbstorage.DB.QueryContext(ctx, smtp)
	if err != nil {
		return api.MetricsMap{}, fmt.Errorf("error when execute select %w", err)
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			logger.Errorf("error when close rows %w", err)
		}
	}()

	for rows.Next() {
		var name string
		var mtype string
		var val interface{}

		err = rows.Scan(&name, &val, &mtype)
		if err != nil {
			return api.MetricsMap{}, fmt.Errorf("error scan %w", err)
		}
		if mtype == api.Gauge {
			value, ok := val.(float64)
			if !ok {
				return api.MetricsMap{}, fmt.Errorf("mismatch metric %s and value type", name)
			}
			metrics.Metrics[name] = api.Metrics{ID: name, MType: mtype, Value: &value}
		}
		if mtype == api.Counter {
			deltaFLoat, ok := val.(float64)
			if !ok {
				return api.MetricsMap{}, fmt.Errorf("mismatch metric %s and delta type", name)
			}
			delta := int64(deltaFLoat)
			metrics.Metrics[name] = api.Metrics{ID: name, MType: mtype, Delta: &delta}
		}
	}

	err = rows.Err()
	if err != nil {
		return api.MetricsMap{}, fmt.Errorf("errors rows %w", err)
	}
	return metrics, nil
}

func (dbstorage *DBStorage) CreateTables(ctx context.Context, logger zap.SugaredLogger) error {
	var err error

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	stmtGauge := `CREATE TABLE IF NOT EXISTS gauge 
	("id" SERIAL PRIMARY KEY, "name" TEXT UNIQUE, "value" FLOAT8)`
	stmtCounter := `CREATE TABLE IF NOT EXISTS counter 
	("id" SERIAL PRIMARY KEY, "name" TEXT UNIQUE, "delta" BIGINT)`
	_, err = dbstorage.DB.ExecContext(ctx, stmtGauge)
	if err != nil {
		return fmt.Errorf("create table error - %w", err)
	}
	_, err = dbstorage.DB.ExecContext(ctx, stmtCounter)
	if err != nil {
		return fmt.Errorf("create table error - %w", err)
	}
	return nil
}

func (dbstorage *DBStorage) UpdateParam(ctx context.Context, cntSummed bool,
	metricType, metricName string, metricValue interface{}, logger zap.SugaredLogger) error {
	stmtGauge := `
	INSERT INTO gauge (name, value) 
	VALUES ($1, $2)
	ON CONFLICT (name) DO UPDATE 
	  SET value = $2`

	stmtCounter := `
	INSERT INTO counter (name, delta) 
	VALUES ($1, $2)
	ON CONFLICT (name) DO UPDATE 
	  SET delta = (SELECT delta FROM counter WHERE name=$1) + $2`

	switch {
	case metricType == api.Gauge:
		err := dbstorage.insertData(ctx, stmtGauge, metricType, metricName, metricValue)
		if err != nil {
			return fmt.Errorf("gauge %w", err)
		}
	case metricType == api.Counter:
		err := dbstorage.insertData(ctx, stmtCounter, metricType, metricName, metricValue)
		if err != nil {
			return fmt.Errorf("counter %w", err)
		}
	default:
		return errors.New("wrong metric type")
	}
	return nil
}

func (dbstorage *DBStorage) GetCounterMetric(ctx context.Context, metricID string,
	logger zap.SugaredLogger) (int64, error) {
	var delta int64
	smtp := `SELECT delta FROM counter WHERE name=$1`

	row := dbstorage.DB.QueryRowContext(ctx, smtp, metricID)

	err := row.Scan(&delta)
	if err != nil {
		return 0, fmt.Errorf("get value counter %s error %v", metricID, err)
	}
	return delta, nil
}

func (dbstorage *DBStorage) GetGaugeMetric(ctx context.Context, metricID string,
	logger zap.SugaredLogger) (float64, error) {
	var value float64
	smtp := `SELECT value FROM gauge WHERE name=$1`

	row := dbstorage.DB.QueryRowContext(ctx, smtp, metricID)

	err := row.Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("get value gauge %s error %v", metricID, err)
	}
	return value, nil
}
