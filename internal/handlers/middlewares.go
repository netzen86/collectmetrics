package handlers

import (
	"fmt"
	"net/http"
	"net/netip"
	"time"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/logger"
)

// WithLogging функция для включения логирования запросов
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		sugar, err := logger.Logger()
		if err != nil {
			http.Error(w, fmt.Sprintf("%v %v\n",
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError), 500)
			return
		}

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
	return http.HandlerFunc(logFn)
}

// AccecsList функция ограничевает доступ к серверу по IP адресу
func AccecsList(network netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ipAddrStr := r.Header.Get(api.ACLHeader)
			switch {
			case len(ipAddrStr) != 0 && network.IsValid():
				ipAddr, err := netip.ParseAddr(ipAddrStr)
				if err != nil {
					http.Error(w, fmt.Sprintf("%v\n",
						http.StatusText(http.StatusInternalServerError)),
						http.StatusInternalServerError)
					return
				}
				if !network.Contains(ipAddr) {
					http.Error(w, fmt.Sprintf("%v\n",
						http.StatusText(http.StatusForbidden)), http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
