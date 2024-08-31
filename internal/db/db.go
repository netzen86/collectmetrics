package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/netzen86/collectmetrics/internal/utils"
)

const (
	DataBaseConString string = "postgres://postgres:collectmetrics@localhost/collectmetrics?sslmode=disable"
)

type ConParam struct {
	Host     string `default:"localhost"`
	User     string `default:"postgres"`
	Password string `default:"collectmetrics"`
	DBname   string `default:"collectmetrics"`
	SSLmode  string `default:"sslmode=disable"`
}

func ConDB(dbconstring string) (*sql.DB, error) {
	db, err := sql.Open("pgx", fmt.Sprint(dbconstring))
	if err != nil {
		return nil, err
	}
	// defer db.Close()
	return db, nil
}

func TableExist(ctx context.Context, tablename, dbconstr string) (bool, error) {
	db, err := ConDB(dbconstr)
	if err != nil {
		return false, err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)

	row := db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename  = ?)",
		tablename)
	defer cancel()

	// готовим переменную для чтения результата
	var value bool
	err = row.Scan(&value) // разбираем результат
	if err != nil {
		return false, err
	}
	return value, nil
}

func CreateTables(ctx context.Context, dbconstr string) error {
	db, err := ConDB(dbconstr)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	stmtGauge := `CREATE TABLE IF NOT EXISTS gauge 
	("id" SERIAL PRIMARY KEY, "name" TEXT UNIQUE, "value" FLOAT8)`
	stmtCounter := `CREATE TABLE IF NOT EXISTS counter 
	("id" SERIAL PRIMARY KEY, "name" TEXT UNIQUE, "delta" BIGINT)`
	_, err = db.ExecContext(ctx, stmtGauge)
	if err != nil {
		return fmt.Errorf("create table error - %w", err)
	}
	_, err = db.ExecContext(ctx, stmtCounter)
	if err != nil {
		return fmt.Errorf("create table error - %w", err)
	}
	return nil
}

func UpdateParamDB(ctx context.Context, dbconstr, metricType, metricName string, metricValue interface{}) error {
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

	db, err := ConDB(dbconstr)
	if err != nil {
		return err
	}
	defer db.Close()

	switch {
	case metricType == "gauge":

		val, err := utils.ParseValGag(metricValue)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, stmtGauge, metricName, val)
		// log.Println("Inserting gauge table value: ", val, "ResVAl: ", resval)
		if err != nil {
			return fmt.Errorf("insert table error - %w", err)
		}

	case metricType == "counter":
		del, err := utils.ParseValCnt(metricValue)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, stmtCounter, metricName, del)
		// log.Println("Inserting counter table value: ", del, "ResVAl: ")
		if err != nil {
			return fmt.Errorf("insert table error - %w", err)
		}
	default:
		return errors.New("wrong metric type")
	}
	return nil
}
