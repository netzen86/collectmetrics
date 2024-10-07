package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
)

const (
	DataBaseConString string = "postgres://postgres:collectmetrics@localhost/collectmetrics?sslmode=disable"
)

// type ConParam struct {
// 	Host     string `default:"localhost"`
// 	User     string `default:"postgres"`
// 	Password string `default:"collectmetrics"`
// 	DBname   string `default:"collectmetrics"`
// 	SSLmode  string `default:"sslmode=disable"`
// }

type DBStorage struct {
	DBconstring string
	DB          *sql.DB
}

func (dbstorage *DBStorage) GetStorage(ctx context.Context) (*memstorage.MemStorage, error) {
	return nil, nil
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

func (dbstorage *DBStorage) CreateTables(ctx context.Context) error {
	var err error
	// db, err := NewDBStorage(ctx, dbconstring)
	// if err != nil {
	// 	return fmt.Errorf("cannot connect fo data base %w", err)
	// }
	// defer db.DB.Close()
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

	// func UpdateParamDB(ctx context.Context, dbconstr, metricType, metricName string, metricValue interface{}) error {
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

	// db, err := NewDBStorage(ctx, dbstorage.DBconstring)
	// if err != nil {
	// 	return fmt.Errorf("cannot connect fo data base %w", err)
	// }

	// defer dbstorage.DB.Close()

	switch {
	case metricType == "gauge":

		val, err := utils.ParseValGag(metricValue)
		if err != nil {
			return err
		}
		_, err = dbstorage.DB.ExecContext(ctx, stmtGauge, metricName, val)
		// log.Println("Inserting gauge table value: ", val, "ResVAl: ", err)
		if err != nil {
			// dbstorage.DB.Close()
			return fmt.Errorf("insert in gauge table error - %w", err)
		}

	case metricType == "counter":
		del, err := utils.ParseValCnt(metricValue)
		if err != nil {
			return err
		}
		_, err = dbstorage.DB.ExecContext(ctx, stmtCounter, metricName, del)
		// log.Println("Inserting counter table value: ", del, "ResVAl:", err)
		if err != nil {
			// dbstorage.DB.Close()
			return fmt.Errorf("insert in counter table error - %w", err)
		}
	default:
		return errors.New("wrong metric type")
	}
	return nil
}

func (dbstorage *DBStorage) GetCounterMetric(ctx context.Context, metricID string) (int64, error) {
	var delta int64
	smtp := `SELECT delta FROM counter WHERE name=$1`

	// defer dbstorage.DB.Close()
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

	// defer dbstorage.DB.Close()
	row := dbstorage.DB.QueryRowContext(ctx, smtp, metricID)

	err := row.Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("get value gauge %s error %v", metricID, err)
	}
	return value, nil
}
