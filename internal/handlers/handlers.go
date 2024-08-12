package handlers

import (
	"context"
	"fmt"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/internal/repositories"
)

func UpdateMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		mValue := chi.URLParam(r, "mValue")
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")

		err := storage.UpdateParam(ctx, mType, mName, mValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Println(storage)
	}
}

func RetrieveMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		t, _ := template.ParseFiles("../../web/template/metrics.html") //setp 1
		t.Execute(w, storage.GetMemStorage(ctx))
		// t.Execute(w, storage.GetMemStorage(ctx).Gauge)
		fmt.Println(storage)
	}
}
