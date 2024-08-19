package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/internal/handlers"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

func main() {
	var endpoint string
	flag.StringVar(&endpoint, "a", "localhost:8080", "Used to set the address and port on which the server runs.")
	flag.Parse()

	endpointTMP := os.Getenv("ADDRESS")
	if len(endpointTMP) != 0 {
		endpoint = endpointTMP
	}

	if len(flag.Args()) != 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	gw := chi.NewRouter()
	memSto, errm := memstorage.NewMemStorage()
	gw.Route("/", func(gw chi.Router) {
		gw.Post("/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update", handlers.WithLogging(handlers.JsonUpdateMHandle(memSto)))
		gw.Post("/value", handlers.WithLogging(handlers.JsonRetrieveOneHandle(memSto)))
		gw.Post("/update/{mType}/{mName}", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.WithLogging(handlers.UpdateMHandle(memSto)))
		gw.Post("/*", handlers.WithLogging(handlers.NotFound))

		gw.Get("/value/{mType}/{mName}", handlers.WithLogging(handlers.RetrieveOneMHandle(memSto)))
		gw.Get("/", handlers.WithLogging(handlers.RetrieveMHandle(memSto)))
		gw.Get("/*", handlers.WithLogging(handlers.NotFound))

	},
	)
	errs := http.ListenAndServe(endpoint, gw)
	if errm != nil || errs != nil {
		panic(fmt.Sprintf("%s %s", errm, errs))
	}
}
