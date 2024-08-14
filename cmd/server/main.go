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
	if len(flag.Args()) != 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	gw := chi.NewRouter()
	memSto, errm := memstorage.NewMemStorage()
	gw.Route("/", func(gw chi.Router) {
		gw.Post("/update/{mType}/{mName}", handlers.BadRequest)
		gw.Post("/update/{mType}/{mName}/", handlers.BadRequest)
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.UpdateMHandle(memSto))
		gw.Get("/value/{mType}/{mName}", handlers.RetrieveOneMHandle(memSto))
		gw.Get("/", handlers.RetrieveMHandle(memSto))
	},
	)
	errs := http.ListenAndServe(endpoint, gw)
	if errm != nil || errs != nil {
		panic(fmt.Sprintf("%s %s", errm, errs))
	}
}
