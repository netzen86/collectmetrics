package db

import (
	"database/sql"
	"fmt"
	"log"

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
	log.Println("in New DB ", dbconstring)

	db, err := sql.Open("pgx", fmt.Sprint(dbconstring))
	if err != nil {
		return nil, err
	}
	// defer db.Close()
	return db, nil
}
