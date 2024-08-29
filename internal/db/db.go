package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

func NewDB(dbconstring string) (*sql.DB, error) {
	db, err := sql.Open("pgx", fmt.Sprint(dbconstring))
	if err != nil {
		return nil, err
	}
	// defer db.Close()
	return db, nil
}

func TableExist(tablename, dbconstr string) (bool, error) {
	db, err := NewDB(dbconstr)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

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

func CreateTables(dbconstr string, tablename ...string) error {
	db, err := NewDB(dbconstr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, tn := range tablename {
		log.Println(tn)
		_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS ?()", tn)
		if err != nil {
			return err
		}
	}
	return nil
}
