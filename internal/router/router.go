package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/handlers"
)

func GetGateway(cfg config.ServerCfg) chi.Router {

	gw := chi.NewRouter()

	gw.Use(handlers.WithLogging)

	gw.Route("/", func(gw chi.Router) {
		gw.Post("/", handlers.BadRequest)
		gw.Post("/update/", handlers.JSONUpdateMMHandle(
			cfg.Storage, cfg.FileStoragePathDef, cfg.SignKeyString, cfg.StoreInterval, cfg.Logger))
		gw.Post("/updates/", handlers.JSONUpdateMMHandle(
			cfg.Storage, cfg.FileStoragePathDef, cfg.SignKeyString, cfg.StoreInterval, cfg.Logger))
		gw.Post("/value/", handlers.JSONRetrieveOneHandle(cfg.Storage, cfg.SignKeyString, cfg.Logger))
		gw.Post("/update/{mType}/{mName}", handlers.BadRequest)
		gw.Post("/update/{mType}/{mName}/", handlers.BadRequest)
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.UpdateMHandle(cfg.Storage, cfg.Logger))
		gw.Post("/*", handlers.NotFound)

		gw.Get("/ping", handlers.PingDB(cfg.DBconstring))
		gw.Get("/value/{mType}/{mName}", handlers.RetrieveOneMHandle(cfg.Storage, cfg.Logger))
		gw.Get("/", handlers.RetrieveMHandle(cfg.Storage))
		gw.Get("/*", handlers.NotFound)
	},
	)
	return gw
}
