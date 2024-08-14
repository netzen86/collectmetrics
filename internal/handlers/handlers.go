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
		mType := chi.URLParam(r, "mType")
		if mType != "counter" && mType != "gauge" {
			http.Error(w, "wrong metric type", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mName := chi.URLParam(r, "mName")
		mValue := chi.URLParam(r, "mValue")
		err := storage.UpdateParam(ctx, mType, mName, mValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func RetrieveMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		t, _ := template.ParseFiles("../../web/template/metrics.html")
		t.Execute(w, storage.GetMemStorage(ctx))
	}
}

func RetrieveOneMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		ctx := context.Background()
		if mType == "counter" {
			counter := storage.GetMemStorage(ctx).Counter
			valueNum, ok := counter[mName]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Write([]byte(fmt.Sprintf("%d", valueNum)))
			return
		}
		if mType == "gauge" {
			gauge := storage.GetMemStorage(ctx).Gauge
			valueNum, ok := gauge[mName]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			valueStr := fmt.Sprintf("%g", valueNum)
			w.Write([]byte(valueStr))
			return
		}
	}
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
