package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/internal/handlers"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

func main() {
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

	errs := http.ListenAndServe(":8080", gw)
	if errm != nil || errs != nil {
		panic(fmt.Sprintf("%s %s", errm, errs))
	}
}
