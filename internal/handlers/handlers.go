package handlers

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/logger"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/db"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/security"
	"github.com/netzen86/collectmetrics/internal/utils"
)

func UpdateMHandle(storage repositories.Repo,
	tempfilename, saveMetricsDefaultPath, dbconstr, storageSelecter string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "mType")
		if mType != "counter" && mType != "gauge" {
			http.Error(w, "wrong metric type", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mName := chi.URLParam(r, "mName")
		mValue := chi.URLParam(r, "mValue")

		// if storageSelecter == "FILE" {
		// 	var tmpmetric api.Metrics
		// 	tmpmetric.MType = mType
		// 	tmpmetric.ID = mName

		// 	if tmpmetric.MType == "counter" {
		// 		tmpDel, err := strconv.ParseInt(mValue, 10, 64)
		// 		if err != nil {
		// 			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
		// 			return
		// 		}
		// 		tmpmetric.Delta = &tmpDel
		// 	} else if tmpmetric.MType == "gauge" {
		// 		tmpVal, err := strconv.ParseFloat(mValue, 64)
		// 		if err != nil {
		// 			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
		// 			return
		// 		}
		// 		tmpmetric.Value = &tmpVal
		// 	} else {
		// 		http.Error(w, fmt.Sprintf("wrong metric type%s\n", http.StatusText(400)), 400)
		// 		return
		// 	}
		// 	// err := storage.FileStorage(r.Context(), tempfilename, tmpmetric)
		// 	// if err != nil {
		// 	// 	http.Error(w, fmt.Sprintf("error in store metric in file %s %v\n", http.StatusText(400), err), 400)
		// 	// 	return
		// 	// }
		// }

		retrybuilder := func() func() error {
			return func() error {
				err := storage.UpdateParam(r.Context(), false, mType, mName, mValue)
				if err != nil {
					log.Println(err)
				}
				return err
			}
		}
		err := utils.RetrayFunc(retrybuilder)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// функция выводит имена метрик хранящихся в хранилище
func RetrieveMHandle(storage repositories.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		if r == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		ctx := r.Context()
		workDir := utils.WorkingDir()
		if !utils.ChkFileExist(workDir + api.TemplatePath) {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles(workDir + api.TemplatePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n",
				http.StatusText(http.StatusInternalServerError), err),
				http.StatusInternalServerError)
			return
		}
		storage, err := storage.GetStorage(ctx)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		t.Execute(&buf, storage)
		data, err := utils.CoHTTP(buf.Bytes(), r, w)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", api.HTML)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// функция для получения значения метрики с помощью URI
func RetrieveOneMHandle(storage repositories.Repo, filename, dbconstr, storageSelecter string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric api.Metrics
		ctx := r.Context()
		metric.MType = chi.URLParam(r, "mType")
		metric.ID = chi.URLParam(r, "mName")

		if metric.MType == "counter" {
			var delta int64
			metric.Delta = &delta
			retrybuilder := func() func() error {
				return func() error {
					var err error
					*metric.Delta, err = storage.GetCounterMetric(ctx, metric.ID)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s with error %v\n",
					http.StatusText(http.StatusNotFound),
					metric.ID, metric.MType, err), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			deltaStr := fmt.Sprintf("%d", *metric.Delta)
			w.Write([]byte(deltaStr))
			return
		} else if metric.MType == "gauge" {
			var value float64
			metric.Value = &value
			retrybuilder := func() func() error {
				return func() error {
					var err error
					*metric.Value, err = storage.GetGaugeMetric(ctx, metric.ID)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s with error %v\n",
					http.StatusText(http.StatusNotFound),
					metric.ID, metric.MType, err), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			valueStr := fmt.Sprintf("%g", *metric.Value)
			w.Write([]byte(valueStr))
			return
		} else {
			http.Error(w, fmt.Sprintf("%s wrong type metric\n", http.StatusText(http.StatusBadRequest)),
				http.StatusBadRequest)
			return
		}
	}
}

// Multiple value update handle
func JSONUpdateMMHandle(storage repositories.Repo,
	tempfilename, filename, dbconstr, storageSelecter, signKey string, time int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var metrics []api.Metrics
		var metric api.Metrics
		var buf bytes.Buffer
		var resp []byte
		var err error

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

		if len(signKey) != 0 && len(r.Header.Get("HashSHA256")) != 0 {
			calcSign := security.SignSendData(buf.Bytes(), []byte(signKey))
			recivedSign, err := hex.DecodeString(r.Header.Get("HashSHA256"))
			if err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(500), "can't decode sign str to []byte"), 500)
				return
			}
			comp := security.CompareSign(calcSign, recivedSign)
			if !comp {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(http.StatusBadRequest), "signature discrepancy"), http.StatusBadRequest)
				return
			}
		}
		// распаковываем если контент упакован
		err = utils.SelectDeCoHTTP(&buf, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), "can't unpack data"), 400)
			return
		}

		if strings.Contains(r.RequestURI, "/updates/") {
			// десериализуем JSON в metrics
			if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), "decode slice metrics to json error"), 400)
				return
			}
		} else if strings.Contains(r.RequestURI, "/update/") {
			// десериализуем JSON в metric
			if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
				log.Println("decode metric error")
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), "decode one metric to json error"), 400)
				return
			}
			metrics = append(metrics, metric)
		}

		for _, metr := range metrics {
			err = MetricParseSelecStor(ctx, storage, &metr, storageSelecter, dbconstr, tempfilename)
			if err != nil {
				log.Println("ERROR", err)
				http.Error(w, fmt.Sprintf("%s %s %v", http.StatusText(400), "can't select sorage ", err), 400)
				return
			}
		}

		if len(metrics) == 0 {
			http.Error(w, "Metrics slice empty try update endpoint", 400)
			return
		}

		switch {
		case r.RequestURI == "/updates/":
			resp, err = json.Marshal(metrics)
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
		case r.RequestURI == "/update/":
			resp, err = json.Marshal(metrics[0])
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}

		if time == 0 {
			newStorage, err := storage.GetStorage(ctx)
			if err != nil {
				http.Error(w, http.StatusText(500), 500)
				return
			}
			var metricsMap api.MetricsMap
			metricsMap.Metrics = make(map[string]api.Metrics)
			memstorage.MemstoragetoMetricMap(&metricsMap, *newStorage)
			err = files.SyncSaveMetrics(metricsMap, filename)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		// packing content
		resp, err = utils.CoHTTP(resp, r, w)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		if len(signKey) != 0 {
			sign := security.SignSendData(resp, []byte(signKey))
			w.Header().Add("HashSHA256", hex.EncodeToString(sign))
		}

		w.Header().Set("Content-Type", api.Js)
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func MetricParseSelecStor(ctx context.Context, storage repositories.Repo,
	metric *api.Metrics, storageSelecter, dbconstr, tempfilename string) error {

	if metric.ID == "" {
		return fmt.Errorf("%s", "not valid metric name")
	}

	if metric.MType == "counter" {
		if metric.Delta == nil {
			return fmt.Errorf("delta nil %s %s", metric.ID, metric.MType)

		}
		// if storageSelecter == "FILE" {
		// 	err := files.FileStorage(ctx, tempfilename, *metric)
		// 	if err != nil {
		// 		return fmt.Errorf("error in store metric counter in file %v", err)
		// 	}
		// } else {
		retrybuilder := func() func() error {
			return func() error {
				err := storage.UpdateParam(ctx, false, metric.MType, metric.ID, *metric.Delta)
				if err != nil {
					log.Println("retry", err)
				}
				return err
			}
		}
		err := utils.RetrayFunc(retrybuilder)
		if err != nil {
			return fmt.Errorf("can't update storage counter value %w", err)
		}

		*metric.Delta, err = storage.GetCounterMetric(ctx, metric.ID)
		if err != nil {
			return fmt.Errorf("can't get updated counter value %w", err)
		}
		// }
	} else if metric.MType == "gauge" {
		if metric.Value == nil {
			return fmt.Errorf("value nil %s %s", metric.ID, metric.MType)
		}

		// if storageSelecter == "FILE" {
		// 	err := files.FileStorage(ctx, tempfilename, *metric)
		// 	if err != nil {
		// 		return fmt.Errorf("error in store metric gauge in file %v", err)
		// 	}
		// } else {
		retrybuilder := func() func() error {
			return func() error {
				err := storage.UpdateParam(ctx, false, metric.MType, metric.ID, *metric.Value)
				if err != nil {
					log.Println("retry", err)
				}
				return err
			}
		}
		err := utils.RetrayFunc(retrybuilder)
		if err != nil {
			return fmt.Errorf("can't update storage gauge value %w", err)
		}

		*metric.Value, err = storage.GetGaugeMetric(ctx, metric.ID)
		if err != nil {
			return fmt.Errorf("can't get updated gauge value %w", err)
		}
		// }
	} else {
		return fmt.Errorf("%s", "empty metic")
	}
	return nil
}

func JSONRetrieveOneHandle(storage repositories.Repo, filename, dbconstr, storageSelecter, signkeystr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var metric *api.Metrics
		var buf bytes.Buffer
		var err error

		// читаем тело запроса
		_, err = buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		// десериализуем JSON в metrics
		if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		if metric.MType == "counter" {
			var delta int64
			metric.Delta = &delta
			retrybuilder := func() func() error {
				return func() error {
					var err error
					*metric.Delta, err = storage.GetCounterMetric(ctx, metric.ID)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s with error %v\n",
					http.StatusText(http.StatusNotFound),
					metric.ID, metric.MType, err), http.StatusNotFound)
				return
			}
		} else if metric.MType == "gauge" {
			var value float64
			metric.Value = &value
			retrybuilder := func() func() error {
				return func() error {
					var err error
					*metric.Value, err = storage.GetGaugeMetric(ctx, metric.ID)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s with error %v\n",
					http.StatusText(http.StatusNotFound),
					metric.ID, metric.MType, err), http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, fmt.Sprintf("%s wrong type metric\n", http.StatusText(400)), 400)
			return
		}
		resp, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		if len(signkeystr) != 0 {
			sign := security.SignSendData(resp, []byte(signkeystr))
			w.Header().Add("HashSHA256", hex.EncodeToString(sign))
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
		ctx := r.Context()
		dbstorage, err := db.NewDBStorage(ctx, dbconstring)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n", http.StatusText(500), err), 500)
			return
		}
		defer dbstorage.DB.Close()

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := dbstorage.DB.PingContext(ctx); err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n", http.StatusText(500), err), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sugar := logger.Loger()

		// функция Now() возвращает текущее время
		start := time.Now()

		// эндпоинт
		uri := r.RequestURI
		// метод запроса
		method := r.Method

		lw, rd := logger.NewLRW(w)

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
