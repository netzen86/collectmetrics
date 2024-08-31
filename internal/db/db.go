package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
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
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	stmtGauge := `CREATE TABLE IF NOT EXISTS gauge 
	("id" SERIAL PRIMARY KEY, "name" TEXT, "value" FLOAT8)`
	stmtCounter := `CREATE TABLE IF NOT EXISTS counter 
	("id" SERIAL PRIMARY KEY, "name" TEXT, "delta" BIGINT)`
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

// func UpdateParamDB(ctx context.Context, dbconstr, metricType, metricName string, metricValue interface{}) error {
// 	return nil
// }
