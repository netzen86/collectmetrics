package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
)

const (
	DataBaseConString string = "postgres://postgres:collectmetrics@localhost/collectmetrics?sslmode=disable"
)

type DBStorage struct {
	DBconstring string
	DB          *sql.DB
}

// функция подключения к базе данных, param = строка для подключения к БД
func NewDBStorage(ctx context.Context, param string) (*DBStorage, error) {
	var dbstorage DBStorage
	var err error
	dbstorage.DB, err = sql.Open("pgx", param)
	if err != nil {
		return &DBStorage{}, fmt.Errorf("cannot connect fo data base %w", err)
	}
	return &dbstorage, nil
}

// функция для вставки данных в базу данных
func insertData(ctx context.Context, dbstorage *DBStorage,
	stmt, metricType, metricName string, metricValue interface{}) error {

	if metricType == api.Gauge {
		val, err := utils.ParseValGag(metricValue)
		if err != nil {
			return err
		}
		_, err = dbstorage.DB.ExecContext(ctx, stmt, metricName, val)
		if err != nil {
			return fmt.Errorf("insert in table error - %w", err)
		}
	}

	if metricType == api.Counter {
		val, err := utils.ParseValCnt(metricValue)
		if err != nil {
			return err
		}
		_, err = dbstorage.DB.ExecContext(ctx, stmt, metricName, val)
		if err != nil {
			return fmt.Errorf("insert in table error - %w", err)
		}
	}
	return nil
}

func (dbstorage *DBStorage) GetStorage(ctx context.Context) (*memstorage.MemStorage, error) {
	return nil, nil
}

func (dbstorage *DBStorage) GetAllMetrics(ctx context.Context) (api.MetricsSlice, error) {
	var metrics api.MetricsSlice

	smtp := `
	SELECT name, value, 'gauge' as type
	FROM gauge
	UNION all
	SELECT name, delta, 'counter' as type
	FROM counter;`

	rows, err := dbstorage.DB.QueryContext(ctx, smtp)
	if err != nil {
		return api.MetricsSlice{}, fmt.Errorf("error when execute select %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var mtype string
		var val interface{}

		err = rows.Scan(&name, &val, &mtype)
		if err != nil {
			return api.MetricsSlice{}, fmt.Errorf("error scan %w", err)
		}
		if mtype == "gauge" {
			value, ok := val.(float64)
			if !ok {
				return api.MetricsSlice{}, fmt.Errorf("mismatch metric %s and value type", name)
			}
			metrics.Metrics = append(metrics.Metrics, api.Metrics{ID: name, MType: mtype, Value: &value})
		}
		if mtype == "counter" {
			delta, ok := val.(float64)
			if !ok {
				return api.MetricsSlice{}, fmt.Errorf("mismatch metric %s and delta type", name)
			}
			metrics.Metrics = append(metrics.Metrics, api.Metrics{ID: name, MType: mtype, Value: &delta})
		}
	}

	err = rows.Err()
	if err != nil {
		return api.MetricsSlice{}, fmt.Errorf("errors rows %w", err)
	}
	return metrics, nil
}

func (dbstorage *DBStorage) CreateTables(ctx context.Context) error {
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

func (dbstorage *DBStorage) UpdateParam(ctx context.Context, cntSummed bool, metricType, metricName string, metricValue interface{}) error {

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
	case metricType == "gauge":
		err := insertData(ctx, dbstorage, stmtGauge, metricType, metricName, metricValue)
		if err != nil {
			return fmt.Errorf("gauge %w", err)
		}
	case metricType == "counter":
		err := insertData(ctx, dbstorage, stmtCounter, metricType, metricName, metricValue)
		if err != nil {
			return fmt.Errorf("counter %w", err)
		}
	default:
		return errors.New("wrong metric type")
	}
	return nil
}

func (dbstorage *DBStorage) GetCounterMetric(ctx context.Context, metricID string) (int64, error) {
	var delta int64
	smtp := `SELECT delta FROM counter WHERE name=$1`

	row := dbstorage.DB.QueryRowContext(ctx, smtp, metricID)

	err := row.Scan(&delta)
	if err != nil {
		return 0, fmt.Errorf("get value counter %s error %v", metricID, err)
	}
	return delta, nil
}

func (dbstorage *DBStorage) GetGaugeMetric(ctx context.Context, metricID string) (float64, error) {
	var value float64
	smtp := `SELECT value FROM gauge WHERE name=$1`

	row := dbstorage.DB.QueryRowContext(ctx, smtp, metricID)

	err := row.Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("get value gauge %s error %v", metricID, err)
	}
	return value, nil
}
