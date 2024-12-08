// Package router - пакет в котором описаны endpoints
package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/config"
	"github.com/netzen86/collectmetrics/internal/handlers"
)

func GetGateway(cfg config.ServerCfg, srvlog zap.SugaredLogger) chi.Router {

	gw := chi.NewRouter()

	gw.Use(handlers.WithLogging)

	gw.Route("/", func(gw chi.Router) {
		gw.Post("/", handlers.BadRequest)
		gw.Post("/update/", handlers.JSONUpdateMMHandle(
			cfg.Storage, cfg.FileStoragePathDef, cfg.SignKeyString,
			cfg.StoreInterval, cfg.PrivKey, srvlog))
		gw.Post("/updates/", handlers.JSONUpdateMMHandle(
			cfg.Storage, cfg.FileStoragePathDef, cfg.SignKeyString,
			cfg.StoreInterval, cfg.PrivKey, srvlog))
		gw.Post("/value/", handlers.JSONRetrieveOneHandle(cfg.Storage, cfg.SignKeyString, srvlog))
		gw.Post("/update/{mType}/{mName}", handlers.BadRequest)
		gw.Post("/update/{mType}/{mName}/", handlers.BadRequest)
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.UpdateMHandle(cfg.Storage, srvlog))
		gw.Post("/*", handlers.NotFound)

		gw.Get("/ping", handlers.PingDB(cfg.DBconstring))
		gw.Get("/value/{mType}/{mName}", handlers.RetrieveOneMHandle(cfg.Storage, srvlog))
		gw.Get("/", handlers.RetrieveMHandle(cfg.Storage, srvlog))
		gw.Get("/*", handlers.NotFound)

		// Define the routes for serving profiling data
		gw.Mount("/debug", middleware.Profiler())
	},
	)
	return gw
}
