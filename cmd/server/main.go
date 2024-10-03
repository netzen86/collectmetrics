package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/handlers"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
)

func main() {
	log.Println("!!! SERVER START !!!")

	cfg, err := config.GetServerCfg()
	if err != nil {
		log.Fatalf("error when getting config %v ", err)
	}

	gw := chi.NewRouter()

	gw.Route("/", func(gw chi.Router) {
		gw.Post("/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/", handlers.WithLogging(handlers.JSONUpdateMMHandle(
			cfg.MemStorage, fmt.Sprintf("%stmp", cfg.FileStoragePath), cfg.FileStoragePathDef,
			cfg.DBconstring, cfg.StorageSelecter, cfg.SignKeyString, cfg.StoreInterval)))
		gw.Post("/updates/", handlers.WithLogging(handlers.JSONUpdateMMHandle(
			cfg.MemStorage, fmt.Sprintf("%stmp", cfg.FileStoragePath), cfg.FileStoragePathDef,
			cfg.DBconstring, cfg.StorageSelecter, cfg.SignKeyString, cfg.StoreInterval)))
		gw.Post("/value/", handlers.WithLogging(handlers.JSONRetrieveOneHandle(
			cfg.MemStorage, fmt.Sprintf("%stmp", cfg.FileStoragePath), cfg.DBconstring,
			cfg.StorageSelecter, cfg.SignKeyString)))
		gw.Post("/update/{mType}/{mName}", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.WithLogging(handlers.UpdateMHandle(
			cfg.MemStorage, fmt.Sprintf("%stmp", cfg.FileStoragePath), cfg.FileStoragePathDef,
			cfg.DBconstring, cfg.StorageSelecter)))
		gw.Post("/*", handlers.WithLogging(handlers.NotFound))

		gw.Get("/ping", handlers.WithLogging(handlers.PingDB(cfg.DBconstring)))
		gw.Get("/value/{mType}/{mName}", handlers.WithLogging(handlers.RetrieveOneMHandle(
			cfg.MemStorage, cfg.FileStoragePath, cfg.DBconstring, cfg.StorageSelecter)))
		gw.Get("/", handlers.WithLogging(handlers.RetrieveMHandle(cfg.MemStorage)))
		gw.Get("/*", handlers.WithLogging(handlers.NotFound))

	},
	)

	if cfg.StoreInterval != 0 {
		go files.SaveMetrics(cfg.MemStorage, cfg.FileStoragePathDef,
			cfg.Tempfile.Name(), cfg.StorageSelecter, cfg.StoreInterval)
	}

	errs := http.ListenAndServe(cfg.Endpoint, gw)
	if errs != nil {
		log.Fatalf("error when start server %v ", errs)
	}
}
