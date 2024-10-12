package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/handlers"
)

func GetGateway(cfg config.ServerCfg) chi.Router {

	gw := chi.NewRouter()

	// gw.Use(handlers.WithLogging)

	gw.Route("/", func(gw chi.Router) {
		gw.Post("/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/", handlers.WithLogging(handlers.JSONUpdateMMHandle(
			cfg.Storage, cfg.FileStoragePathDef, cfg.SignKeyString, cfg.StoreInterval)))
		gw.Post("/updates/", handlers.WithLogging(handlers.JSONUpdateMMHandle(
			cfg.Storage, cfg.FileStoragePathDef, cfg.SignKeyString, cfg.StoreInterval)))
		gw.Post("/value/", handlers.WithLogging(handlers.JSONRetrieveOneHandle(
			cfg.Storage, cfg.SignKeyString)))
		gw.Post("/update/{mType}/{mName}", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.WithLogging(handlers.UpdateMHandle(
			cfg.Storage)))
		gw.Post("/*", handlers.WithLogging(handlers.NotFound))

		gw.Get("/ping", handlers.WithLogging(handlers.PingDB(cfg.DBconstring)))
		gw.Get("/value/{mType}/{mName}", handlers.WithLogging(handlers.RetrieveOneMHandle(
			cfg.Storage)))
		gw.Get("/", handlers.WithLogging(handlers.RetrieveMHandle(cfg.Storage)))
		gw.Get("/*", handlers.WithLogging(handlers.NotFound))

	},
	)
	return gw
}
