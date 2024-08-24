package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/loger"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/utils"
)

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
		var buf bytes.Buffer
		if r == nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		ctx := r.Context()
		workDir := utils.WorkingDir()
		log.Println("work dir:", workDir, "template: ", api.TemplatePath)
		if !utils.ChkFileExist(workDir + api.TemplatePath) {
			http.Error(w, http.StatusText(http.StatusTeapot), http.StatusTeapot)
			return
		}
		t, err := template.ParseFiles(workDir + api.TemplatePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n", http.StatusText(500), err), 500)
			return
		}
		storage, err := storage.GetMemStorage(ctx)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		t.Execute(&buf, storage)
		data, err := utils.CoHTTP(buf.Bytes(), r, w)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func RetrieveOneMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		mName := chi.URLParam(r, "mName")
		ctx := r.Context()
		if mType == "counter" {
			storage, err := storage.GetMemStorage(ctx)
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
			counter := storage.Counter
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
			storage, err := storage.GetMemStorage(ctx)
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
			gauge := storage.Gauge
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

func JSONUpdateMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var metrics api.Metrics
		var buf bytes.Buffer
		newStorage, err := storage.GetMemStorage(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(500), "error create new storage"), 500)
			return
		}

		// читаем тело запроса
		readedbytes, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), "error body data reading"), 400)
			return
		}

		// отвечаем агенту что поддерживаем компрессию
		if strings.Contains(r.Header.Get("Content-Encoding"), api.Gz) && readedbytes == 0 {
			w.Header().Add("Accept-Encoding", api.Gz)
			w.WriteHeader(http.StatusOK)
			return
		}

		// распаковываем если контент упакован
		err = utils.SelectDeCoHTTP(&buf, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), "can't unpack data"), 400)
			return
		}

		// десериализуем JSON в metrics
		if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), "decode to json error"), 400)
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
		// if strings.Contains(r.Header.Get("Accept-Encoding"), api.Gz) {
		// 	resp, err = utils.GzipCompress(resp)
		// 	if err != nil {
		// 		http.Error(w, http.StatusText(500), 500)
		// 		return
		// 	}
		// 	w.Header().Set("Content-Encoding", api.Gz)
		// }

		resp, err = utils.CoHTTP(resp, r, w)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		w.Header().Set("Content-Type", api.Js)
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func JSONRetrieveOneHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var metrics api.Metrics
		var buf bytes.Buffer

		storage, err := storage.GetMemStorage(ctx)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		// читаем тело запроса
		_, err = buf.ReadFrom(r.Body)
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
			value, ok := storage.Counter[metrics.ID]
			if !ok {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s\n",
					http.StatusText(404),
					metrics.ID, metrics.MType), 404)
				return
			}
			metrics.Delta = &value
		}
		if metrics.MType == "gauge" {
			value, ok := storage.Gauge[metrics.ID]
			if !ok {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s\n",
					http.StatusText(404),
					metrics.ID, metrics.MType), 404)
				return
			}
			metrics.Value = &value
		}

		resp, err := json.Marshal(metrics)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		resp, err = utils.CoHTTP(resp, r, w)
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
	http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), "from function"), 400)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), 404)
}
