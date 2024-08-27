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

func NewDB(conparam ConParam) (*sql.DB, error) {
	db, err := sql.Open("pgx", fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s",
		conparam.Host, conparam.User, conparam.Password, conparam.DBname, conparam.SSLmode))
	if err != nil {
		return nil, err
	}
	// defer db.Close()
	return db, nil
}
