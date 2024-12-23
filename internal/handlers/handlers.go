// Package handlers - пакет содержит функции для обработки http запросов
package handlers

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/db"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/security"
	"github.com/netzen86/collectmetrics/internal/utils"
)

// UpdateMHandle функция для изменения значений метрик с помощью URI
func UpdateMHandle(storage repositories.Repo, srvlog zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// если тип хранилища файлсторэж то
		// необходимо суммировать метрики типа counter
		_, cntSummed := storage.(*files.Filestorage)

		mType := chi.URLParam(r, "mType")
		if mType != api.Counter && mType != api.Gauge {
			http.Error(w, "wrong metric type", http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mName := chi.URLParam(r, "mName")
		mValue := chi.URLParam(r, "mValue")

		retrybuilder := func() func() error {
			return func() error {
				err := storage.UpdateParam(r.Context(), cntSummed, mType,
					mName, mValue, srvlog)
				if err != nil {
					srvlog.Warnf("error updating from uri %w", err)
				}
				return err
			}
		}
		err := utils.RetryFunc(retrybuilder)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), err), 400)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// RetrieveMHandle функция выводит имена метрик хранящихся в хранилище
func RetrieveMHandle(storage repositories.Repo, srvlog zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		if r == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		ctx := r.Context()

		// получаем корень папки проекта
		workDir := utils.WorkingDir()

		// проверяем существует ли файл жаблона страшицы
		if !utils.ChkFileExist(workDir + api.TemplatePath) {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		// создаем шаблон
		t, err := template.ParseFiles(workDir + api.TemplatePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n",
				http.StatusText(http.StatusInternalServerError), err),
				http.StatusInternalServerError)
			return
		}
		// получаем метрики
		metrics, err := storage.GetAllMetrics(ctx, srvlog)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		// генерим шаблон
		err = t.Execute(&buf, metrics)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		data, err := utils.CoHTTP(buf.Bytes(), r, w)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", api.HTML)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
	}
}

// RetrieveOneMHandle функция для получения значения метрики с помощью URI
func RetrieveOneMHandle(storage repositories.Repo, srvlog zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric api.Metrics
		ctx := r.Context()
		metric.MType = chi.URLParam(r, "mType")
		metric.ID = chi.URLParam(r, "mName")

		switch {
		case metric.MType == api.Counter:
			var delta int64
			metric.Delta = &delta
			retrybuilder := func() func() error {
				return func() error {
					var err error
					*metric.Delta, err = storage.GetCounterMetric(ctx, metric.ID, srvlog)
					if err != nil {
						srvlog.Infof("error getting metirc %s - %w", metric.ID, err)
					}
					return err
				}
			}
			err := utils.RetryFunc(retrybuilder)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s with error %v\n",
					http.StatusText(http.StatusNotFound),
					metric.ID, metric.MType, err), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			deltaStr := fmt.Sprintf("%d", *metric.Delta)
			_, err = w.Write([]byte(deltaStr))
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
			return

		case metric.MType == api.Gauge:
			var value float64
			metric.Value = &value
			retrybuilder := func() func() error {
				return func() error {
					var err error
					*metric.Value, err = storage.GetGaugeMetric(ctx, metric.ID, srvlog)
					if err != nil {
						srvlog.Infof("error getting metirc %s - %w", metric.ID, err)
					}
					return err
				}
			}
			err := utils.RetryFunc(retrybuilder)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s with error %v\n",
					http.StatusText(http.StatusNotFound),
					metric.ID, metric.MType, err), http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			valueStr := fmt.Sprintf("%g", *metric.Value)
			_, err = w.Write([]byte(valueStr))
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
			return
		default:
			http.Error(w, fmt.Sprintf("%s wrong type metric\n", http.StatusText(http.StatusBadRequest)),
				http.StatusBadRequest)
			return
		}
	}
}

// JSONUpdateMMHandle хэндлер для обработки нескольких запросов
func JSONUpdateMMHandle(storage repositories.Repo, filename,
	signKey string, time int, privKey *rsa.PrivateKey, srvlog zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var metrics []api.Metrics
		var metricsMap api.MetricsMap
		var metric api.Metrics
		var buf bytes.Buffer
		var recivedSign []byte
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

		// проверяем подпись в заголовке
		if len(signKey) != 0 && len(r.Header.Get("HashSHA256")) != 0 {
			calcSign := security.SignSendData(buf.Bytes(), []byte(signKey))
			recivedSign, err = hex.DecodeString(r.Header.Get("HashSHA256"))
			if err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(http.StatusInternalServerError), "can't decode sign str to []byte"),
					http.StatusInternalServerError)
				return
			}
			comp := security.CompareSign(calcSign, recivedSign)
			if !comp {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(http.StatusBadRequest), "signature discrepancy"),
					http.StatusBadRequest)
				return
			}
		}

		// распаковываем если контент упакован
		err = utils.SelectDeCoHTTP(&buf, r, srvlog)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(http.StatusBadRequest),
				"can't unpack data"), http.StatusBadRequest)
			return
		}

		// расшифровываем если контент зашифрован
		if len(r.Header.Get("CryptRSA")) != 0 {
			err = security.DecryptMetric(&buf, privKey)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(http.StatusInternalServerError),
					"can't decrypt data"), http.StatusInternalServerError)
				return
			}
		}

		// десериализуем JSON в metrics
		if strings.Contains(r.RequestURI, "/updates/") {
			if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(http.StatusBadRequest),
					"decode slice metrics to json error"), http.StatusBadRequest)
				return
			}
		} else if strings.Contains(r.RequestURI, "/update/") {
			if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
				srvlog.Warnf("decode metric %s error %w", metric.ID, err)
				http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(http.StatusBadRequest),
					"decode one metric to json error"), http.StatusBadRequest)
				return
			}
			metrics = append(metrics, metric)
		}

		if len(metrics) == 0 {
			http.Error(w, "Metrics slice empty try update endpoint", 400)
			return
		}

		for _, metric := range metrics {
			err = MetricParseSelecStor(ctx, storage, &metric, srvlog)
			if err != nil {
				srvlog.Warnf("error parse metric %s error %w", metric.ID, err)
				http.Error(w, fmt.Sprintf("%s %s %v", http.StatusText(http.StatusBadRequest),
					"can't select sorage ", err), http.StatusBadRequest)
				return
			}
		}

		switch {
		case r.RequestURI == "/updates/":
			resp, err = json.Marshal(metrics)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
		case r.RequestURI == "/update/":
			resp, err = json.Marshal(metrics[0])
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
		}

		// сохраняем метрики в файл синхронно с запросом
		if time == 0 {
			metricsMap, err = storage.GetAllMetrics(ctx, srvlog)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
			err = files.SyncSaveMetrics(metricsMap, filename, srvlog)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
		}

		// запаковываем контент
		resp, err = utils.CoHTTP(resp, r, w)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}

		// добавляем подпись в заголовок
		if len(signKey) != 0 {
			sign := security.SignSendData(resp, []byte(signKey))
			w.Header().Add("HashSHA256", hex.EncodeToString(sign))
		}

		w.Header().Set("Content-Type", api.Js)
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
	}
}

// MetricParseSelecStor функция для сохраненние метрик в хранилище
func MetricParseSelecStor(ctx context.Context, storage repositories.Repo,
	metric *api.Metrics, srvlog zap.SugaredLogger) error {

	// переменная для определения того нужно ли
	// складывать значение метрики типа counter, используется для файлсторэжа
	_, cntSummed := storage.(*files.Filestorage)

	if metric.ID == "" {
		return fmt.Errorf("%s", "not valid metric name")
	}

	switch {
	case metric.MType == api.Counter:
		if metric.Delta == nil {
			return fmt.Errorf("delta nil %s %s", metric.ID, metric.MType)
		}
		retrybuilder := func() func() error {
			return func() error {
				err := storage.UpdateParam(ctx, cntSummed, metric.MType,
					metric.ID, *metric.Delta, srvlog)
				if err != nil {
					srvlog.Warnf("error updating metric %w", err)
				}
				return err
			}
		}
		err := utils.RetryFunc(retrybuilder)
		if err != nil {
			return fmt.Errorf("can't update storage counter value %w", err)
		}
		*metric.Delta, err = storage.GetCounterMetric(ctx, metric.ID, srvlog)
		if err != nil {
			return fmt.Errorf("can't get updated counter value %w", err)
		}
	case metric.MType == api.Gauge:
		if metric.Value == nil {
			return fmt.Errorf("value nil %s %s", metric.ID, metric.MType)
		}
		retrybuilder := func() func() error {
			return func() error {
				err := storage.UpdateParam(ctx, cntSummed,
					metric.MType, metric.ID, *metric.Value, srvlog)
				if err != nil {
					srvlog.Warnf("error updating metiric %w", err)
				}
				return err
			}
		}
		err := utils.RetryFunc(retrybuilder)
		if err != nil {
			return fmt.Errorf("can't update storage gauge value %w", err)
		}
		*metric.Value, err = storage.GetGaugeMetric(ctx, metric.ID, srvlog)
		if err != nil {
			return fmt.Errorf("can't get updated gauge value %w", err)
		}
	default:
		return fmt.Errorf("%s", "empty metic")
	}
	return nil
}

// JSONRetrieveOneHandle функция для получения одного значения из хранилища
func JSONRetrieveOneHandle(storage repositories.Repo, signkeystr string,
	srvlog zap.SugaredLogger) http.HandlerFunc {
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
		// зарашиваем метрики
		switch {
		case metric.MType == api.Counter:
			var delta int64
			metric.Delta = &delta
			// делаем несколько попыток получить метрику
			retrybuilder := func() func() error {
				return func() error {
					*metric.Delta, err = storage.GetCounterMetric(ctx, metric.ID,
						srvlog)
					if err != nil {
						srvlog.Warnf("error getting metric %w", err)
					}
					return err
				}
			}
			err = utils.RetryFunc(retrybuilder)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s with error %v\n",
					http.StatusText(http.StatusNotFound),
					metric.ID, metric.MType, err), http.StatusNotFound)
				return
			}
		case metric.MType == api.Gauge:
			var value float64
			metric.Value = &value
			retrybuilder := func() func() error {
				return func() error {
					*metric.Value, err = storage.GetGaugeMetric(ctx, metric.ID,
						srvlog)
					if err != nil {
						srvlog.Warnf("error getting metric %w", err)
					}
					return err
				}
			}
			err = utils.RetryFunc(retrybuilder)
			if err != nil {
				http.Error(w, fmt.Sprintf(
					"%s - metric %s not exist in %s with error %v\n",
					http.StatusText(http.StatusNotFound),
					metric.ID, metric.MType, err), http.StatusNotFound)
				return
			}
		default:
			http.Error(w, fmt.Sprintf("%s wrong type metric\n", http.StatusText(400)), 400)
			return
		}
		// сериализуем метрики полученные из хранилища
		resp, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		// добавляем подпись в заголовок
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
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
	}
}

// PingDB функция для проверки подключения к базе данных
func PingDB(dbconstring string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dbstorage, err := db.NewDBStorage(ctx, dbconstring)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n", http.StatusText(500), err), 500)
			return
		}
		defer func() {
			err = dbstorage.DB.Close()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
		}()

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := dbstorage.DB.PingContext(ctx); err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n", http.StatusText(500), err), 500)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("%s %v\n", http.StatusText(400), "from function"), 400)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), 404)
}
