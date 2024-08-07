package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/netzen86/collectmetrics/internal/repositories"
)

func GetMetrics(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		if r.Method == http.MethodPost {
			uri := strings.Split(r.RequestURI, "/")
			if len(uri) == 5 {
				err := storage.UpdateParam(ctx, uri[2], uri[3], uri[4])
				if uri[1] != "update" {
					err = errors.New("wrong method use update")
				}
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					w.WriteHeader(http.StatusBadRequest)
				}
				w.WriteHeader(http.StatusOK)
			} else {
				http.Error(w, "Metrics name not found!", http.StatusNotFound)
				w.WriteHeader(http.StatusNotFound)
			}
			return
		} else {
			http.Error(w, "Use POST method!", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
