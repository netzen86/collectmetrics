// Package logger - пакет для логирования
package logger

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

func Logger() (zap.SugaredLogger, error) {
	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		return zap.SugaredLogger{},
			fmt.Errorf("error when crating logger %w", err)
	}
	defer func() {
		err = logger.Sync()
	}()

	// делаем регистратор SugaredLogger
	return *logger.Sugar(), nil
}

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		Status int
		Size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func NewLRW(w http.ResponseWriter) (loggingResponseWriter, *responseData) {
	rd := &responseData{
		Status: 0,
		Size:   0,
	}
	return loggingResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   rd,
	}, rd
}

func (r loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.Size += size // захватываем размер
	return size, err
}

func (r loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.Status = statusCode // захватываем код статуса
}
