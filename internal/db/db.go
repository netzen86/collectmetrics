package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ConParam struct {
	Host     string `default:"localhost"`
	User     string `default:"postgres"`
	Password string `default:"collectmetrics"`
	DBname   string `default:"collectmetrics"`
	SSLmode  string `default:"sslmode=disable"`
}

func NewDB(dbconstring string) (*sql.DB, error) {
	db, err := sql.Open("pgx", fmt.Sprint("postgres://", dbconstring))
	if err != nil {
		return nil, err
	}
	// defer db.Close()
	return db, nil
}
