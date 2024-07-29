package main

import (
	"fmt"
	"net/http"

	"github.com/netzen86/collectmetrics/internal/handlers"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

func main() {
	memSto, errm := memstorage.NewMemStorage()
	http.HandleFunc("/", handlers.GetMetrics(memSto))
	http.Handle("/update", http.NotFoundHandler())
	http.Handle("/update/counter", http.NotFoundHandler())
	http.Handle("/update/gauge", http.NotFoundHandler())

	errs := http.ListenAndServe(`:8080`, nil)
	if errm != nil || errs != nil {
		panic(fmt.Sprintf("%s %s", errm, errs))
	}
}
