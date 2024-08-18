package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/internal/loger"
	"github.com/netzen86/collectmetrics/internal/repositories"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func UpdateMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
		ctx := r.Context()
		t, _ := template.ParseFiles("../../web/template/metrics.html")
		t.Execute(w, storage.GetMemStorage(ctx))
		w.WriteHeader(http.StatusOK)
	}
}

func RetrieveOneMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		ctx := r.Context()
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

func JsonUpdateMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var metrics Metrics
		var buf bytes.Buffer
		newStorage := storage.GetMemStorage(ctx)
		// читаем тело запроса
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		// десериализуем JSON в metrics
		if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		if metrics.MType == "counter" {
			if metrics.Delta == nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(500), "wrong value"), 500)
				return

			}
			err := storage.UpdateParam(ctx, metrics.MType, metrics.ID, *metrics.Delta)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(500), err), 500)
				return
			}
			*metrics.Delta = newStorage.Counter[metrics.ID]
		} else if metrics.MType == "gauge" {
			if metrics.Value == nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(500), "wrong value"), 500)
				return

			}
			err := storage.UpdateParam(ctx, metrics.MType, metrics.ID, *metrics.Value)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(500), err), 500)
				return
			}
			*metrics.Value = newStorage.Gauge[metrics.ID]
		}

		resp, err := json.Marshal(metrics)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
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

func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(400), 400)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), 404)
}
