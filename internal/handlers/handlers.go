package handlers

import (
	"context"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/internal/loger"
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
		w.WriteHeader(http.StatusOK)
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
			w.WriteHeader(http.StatusOK)
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
			w.WriteHeader(http.StatusOK)
			valueStr := fmt.Sprintf("%g", valueNum)
			w.Write([]byte(valueStr))
			return
		}
	}
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("%d Bad Request", http.StatusBadRequest), http.StatusBadRequest)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("%d Not Found", http.StatusNotFound), http.StatusNotFound)
}

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sugar := loger.Loger()

		// функция Now() возвращает текущее время
		start := time.Now()

		// эндпоинт
		uri := r.RequestURI
		// метод запроса
		method := r.Method

		lw, rd := loger.NewLRW(w)

		// точка, где выполняется хендлер pingHandler
		h.ServeHTTP(lw, r) // обслуживание оригинального запроса

		// Since возвращает разницу во времени между start
		// и моментом вызова Since. Таким образом можно посчитать
		// время выполнения запроса.
		duration := time.Since(start)
		// fmt.Println(uri, method, duration)
		// отправляем сведения о запросе в zap
		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"status", rd.Status,
			"size", rd.Size,
		)
	}
}
