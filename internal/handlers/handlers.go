package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/db"
	"github.com/netzen86/collectmetrics/internal/loger"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
)

func sumPc(ctx context.Context, filename string, delta int64, pcMetric *api.Metrics) error {
	consumer, err := files.NewConsumer(filename)
	if err != nil {
		return err
	}
	_, err = files.ReadOneMetric(ctx, consumer, pcMetric)
	if err != nil {
		return err
	}

	*pcMetric.Delta = *pcMetric.Delta + delta

	return nil
}

func fileStorage(ctx context.Context, pcMetric *api.Metrics, tmpmetric api.Metrics, filename string) error {
	var err error

	tmpStorage, err := memstorage.NewMemStorage()
	if err != nil {
		return err
	}

	if tmpmetric.MType == "counter" {
		pcMetric.ID = tmpmetric.ID
		pcMetric.MType = tmpmetric.MType
		pcMetric.Delta = tmpmetric.Delta

		err = sumPc(ctx, filename, *tmpmetric.Delta, pcMetric)
		if err != nil {
			return err
		}
	}

	files.LoadMetric(tmpStorage, filename)

	if tmpmetric.MType == "counter" {
		tmpStorage.Counter[tmpmetric.ID] = *pcMetric.Delta
	} else if tmpmetric.MType == "gauge" {
		tmpStorage.Gauge[tmpmetric.ID] = *tmpmetric.Value
	}

	files.SyncSaveMetrics(tmpStorage, filename)
	if err != nil {
		return err
	}
	return nil
}

func UpdateMHandle(storage repositories.Repo, pcMetric *api.Metrics, producer *files.Producer,
	dbconstr, storageSelecter string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		if mType != "counter" && mType != "gauge" {
			http.Error(w, "wrong metric type", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mName := chi.URLParam(r, "mName")
		mValue := chi.URLParam(r, "mValue")

		if storageSelecter == "MEMORY" {
			err := storage.UpdateParam(r.Context(), false, mType, mName, mValue)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
				return
			}
		}
		if storageSelecter == "DATABASE" {
			err := db.UpdateParamDB(r.Context(), dbconstr, mType, mName, mValue)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
				return
			}
		}
		if storageSelecter == "FILE" {
			var tmpmetric api.Metrics
			tmpmetric.MType = mType
			tmpmetric.ID = mName

			if tmpmetric.MType == "counter" {
				tmpDel, err := strconv.ParseInt(mValue, 10, 64)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
					return
				}
				tmpmetric.Delta = &tmpDel
			} else if tmpmetric.MType == "gauge" {
				tmpVal, err := strconv.ParseFloat(mValue, 64)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
					return
				}
				tmpmetric.Value = &tmpVal
			} else {
				http.Error(w, fmt.Sprintf("wrong metric type%s\n", http.StatusText(400)), 400)
				return
			}
			err := fileStorage(r.Context(), pcMetric, tmpmetric, producer.Filename)
			if err != nil {
				http.Error(w, fmt.Sprintf("error in store metric in file %s %v\n", http.StatusText(400), err), 400)
				return
			}
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
		if !utils.ChkFileExist(workDir + api.TemplatePath) {
			http.Error(w, http.StatusText(500), 500)
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
		w.Header().Set("Content-Type", api.HTML)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func RetrieveOneMHandle(storage repositories.Repo, filename, dbconstr, storageSelecter string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric api.Metrics
		var value float64
		var delta int64
		var ok bool
		metric.MType = chi.URLParam(r, "mType")
		metric.ID = chi.URLParam(r, "mName")

		if metric.MType == "counter" {
			if storageSelecter == "MEMORY" {
				storage, err := storage.GetMemStorage(r.Context())
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
				delta, ok = storage.Counter[metric.ID]
				if !ok {
					http.Error(w, fmt.Sprintf(
						"%s - metric %s not exist in %s\n",
						http.StatusText(404),
						metric.ID, metric.MType), 404)
					return
				}
				metric.Delta = &delta
			}
			if storageSelecter == "DATABASE" {
				err := db.RetriveOneMetricDB(r.Context(), dbconstr, &metric)
				if err != nil {
					http.Error(w, fmt.Sprintf(
						"%s - metric %s not exist in %s with error %v\n",
						http.StatusText(404),
						metric.ID, metric.MType, err), 404)
					return
				}
			}
			if storageSelecter == "FILE" {
				consumer, err := files.NewConsumer(filename)
				if err != nil {
					http.Error(w, http.StatusText(404), 404)
					return
				}
				_, err = files.ReadOneMetric(r.Context(), consumer, &metric)
				if err != nil {
					http.Error(w, http.StatusText(404), 404)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
			deltaStr := fmt.Sprintf("%d", *metric.Delta)
			w.Write([]byte(deltaStr))
			return
		} else if metric.MType == "gauge" {
			log.Println("DB metric", metric.ID, metric.MType)
			if storageSelecter == "MEMORY" {
				storage, err := storage.GetMemStorage(r.Context())
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
				value, ok = storage.Gauge[metric.ID]
				if !ok {
					http.Error(w, fmt.Sprintf(
						"%s - metric %s not exist in %s\n",
						http.StatusText(404),
						metric.ID, metric.MType), 404)
					return
				}
				metric.Value = &value
			}
			if storageSelecter == "DATABASE" {
				err := db.RetriveOneMetricDB(r.Context(), dbconstr, &metric)

				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
			}
			if storageSelecter == "FILE" {
				consumer, err := files.NewConsumer(filename)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
				_, err = files.ReadOneMetric(r.Context(), consumer, &metric)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}

			}
			w.WriteHeader(http.StatusOK)
			valueStr := fmt.Sprintf("%g", *metric.Value)
			w.Write([]byte(valueStr))
			return
		} else {
			http.Error(w, fmt.Sprintf("%s wrong type metric\n", http.StatusText(400)), 400)
			return
		}
	}
}

func JSONUpdateMHandle(storage repositories.Repo, pcMetric *api.Metrics, producer *files.Producer,
	filename, dbconstr, storageSelecter string, time int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var newStorage *memstorage.MemStorage
		var metrics api.Metrics
		var buf bytes.Buffer
		var err error
		if storageSelecter == "MEMORY" {
			newStorage, err = storage.GetMemStorage(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(500), "error create new storage"), 500)
				return
			}
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
			if storageSelecter == "MEMORY" {
				err := storage.UpdateParam(ctx, false, metrics.MType, metrics.ID, *metrics.Delta)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
					return
				}
				*metrics.Delta = newStorage.Counter[metrics.ID]
			}
			if storageSelecter == "DATABASE" {
				err = db.UpdateParamDB(r.Context(), dbconstr, metrics.MType, metrics.ID, *metrics.Delta)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
					return
				}
			}
			if storageSelecter == "FILE" {
				var tmpmetric api.Metrics
				err := fileStorage(r.Context(), pcMetric, tmpmetric, producer.Filename)
				if err != nil {
					http.Error(w, fmt.Sprintf("error in store metric in file %s %v\n", http.StatusText(400), err), 400)
					return
				}
			}
		} else if metrics.MType == "gauge" {
			if metrics.Value == nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(500), "wrong value"), 500)
				return

			}
			if storageSelecter == "MEMORY" {
				err := storage.UpdateParam(ctx, false, metrics.MType, metrics.ID, *metrics.Value)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
					return
				}
				*metrics.Value = newStorage.Gauge[metrics.ID]
			}
			if storageSelecter == "DATABASE" {
				err = db.UpdateParamDB(r.Context(), dbconstr, metrics.MType, metrics.ID, *metrics.Value)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
					return
				}
			}
			if storageSelecter == "FILE" {
				err := files.UpdateParamFile(r.Context(), producer, metrics.MType, metrics.ID, *metrics.Value)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
					return
				}
			}
		} else {
			http.Error(w, fmt.Sprintf("empty metic %s\n", http.StatusText(400)), 400)
			return
		}

		resp, err := json.Marshal(metrics)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if time == 0 {
			files.SyncSaveMetrics(newStorage, filename)
		}
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

func JSONRetrieveOneHandle(storage repositories.Repo, filename, dbconstr, storageSelecter string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var metrics *api.Metrics
		var buf bytes.Buffer
		var newStorage *memstorage.MemStorage
		var err error

		if storageSelecter == "MEMORY" {
			newStorage, err = storage.GetMemStorage(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"get memstorage %s - %v",
					http.StatusText(500), err), 500)
				return
			}
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
			if storageSelecter == "MEMORY" {
				value, ok := newStorage.Counter[metrics.ID]
				if !ok {
					http.Error(w, fmt.Sprintf(
						"%s - metric %s not exist in %s\n",
						http.StatusText(404),
						metrics.ID, metrics.MType), 404)
					return
				}
				metrics.Delta = &value
			}
			if storageSelecter == "DATABASE" {
				err = db.RetriveOneMetricDB(r.Context(), dbconstr, metrics)
				if err != nil {
					http.Error(w, fmt.Sprintf(
						"%s - metric %s not exist in %s error - %v",
						http.StatusText(404),
						metrics.ID, metrics.MType, err), 404)
					return
				}
			}
			if storageSelecter == "FILE" {
				consumer, err := files.NewConsumer(filename)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
				_, err = files.ReadOneMetric(r.Context(), consumer, metrics)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
			}
		} else if metrics.MType == "gauge" {
			if storageSelecter == "MEMORY" {
				value, ok := newStorage.Gauge[metrics.ID]
				if !ok {
					http.Error(w, fmt.Sprintf(
						"%s - metric %s not exist in %s\n",
						http.StatusText(404),
						metrics.ID, metrics.MType), 404)
					return
				}
				metrics.Value = &value
			}
			if storageSelecter == "DATABASE" {
				err = db.RetriveOneMetricDB(r.Context(), dbconstr, metrics)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
			}
			if storageSelecter == "FILE" {
				consumer, err := files.NewConsumer(filename)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
				_, err = files.ReadOneMetric(r.Context(), consumer, metrics)
				if err != nil {
					http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(404), err), 404)
					return
				}
			}

		} else {
			http.Error(w, fmt.Sprintf("%s wrong type metric\n", http.StatusText(400)), 400)
			return
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

		w.Header().Set("Content-Type", api.Js)
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func PingDB(dbconstring string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("in handler ", dbconstring)

		dataBase, err := db.ConDB(dbconstring)
		if err != nil {
			panic(err)
		}
		defer dataBase.Close()

		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second)
		defer cancel()
		if err := dataBase.PingContext(ctx); err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n", http.StatusText(500), err), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
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
