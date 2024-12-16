// Package utils - пакет содержит вспомогательные функции
package utils

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
)

const (
	backoffII       time.Duration = 1
	backoffRF       float64       = 0.5
	backoffMult     float64       = 2
	backoffMaxETime time.Duration = 9
	NotAval         string        = "N/A"
)

var (
	buildVersion string = NotAval
	buildDate    string = NotAval
	buildCommit  string = NotAval
)

func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func GzipDecompress(buf *bytes.Buffer, logger zap.SugaredLogger) error {
	// переменная buf будет читать входящие данные и распаковывать их
	logger.Infoln("*** DECOMPRESSING ***")
	gz, err := gzip.NewReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("!!!%s!!! unpacking data error", err)
	}
	defer func() {
		err = gz.Close()
		if err != nil {
			logger.Errorf("error when clogse gz %w", err)
		}
	}()

	// в отчищенную переменную buf записываются распакованные данные
	buf.Reset()
	_, err = buf.ReadFrom(gz)
	if err != nil {
		return fmt.Errorf("error when read from gz %w", err)
	}
	return nil
}

func SelectDeCoHTTP(buf *bytes.Buffer, r interface{},
	logger zap.SugaredLogger) error {
	var key string
	switch value := r.(type) {
	case *http.Request:
		key = value.Header.Get("Content-Encoding")
	case *http.Response:
		key = value.Header.Get("Content-Encoding")
	default:
		return errors.New("func get second arg - http.Request/Response")
	}
	if strings.Contains(key, "gzip") {
		err := GzipDecompress(buf, logger)
		if err != nil {
			return fmt.Errorf("!!%s!! unpack data error", err)
		}
	}
	return nil
}

// CoHTTP функция определяем нужно ли сжимать контент если нужно сжимает Gzip'ом
func CoHTTP(data []byte, r *http.Request, w http.ResponseWriter) ([]byte, error) {
	var err error
	if strings.Contains(r.Header.Get("Accept-Encoding"), api.Gz) &&
		(strings.Contains(r.Header.Get("Content-Type"), api.Th) ||
			strings.Contains(r.Header.Get("Content-Type"), api.Js) ||
			strings.Contains(r.Header.Get("Accept"), api.Js) ||
			strings.Contains(r.Header.Get("Accept"), api.Th)) {
		data, err = GzipCompress(data)
		if err != nil {
			return nil, fmt.Errorf("%s pack data error", err)
		}
		w.Header().Set("Content-Encoding", api.Gz)
	}
	return data, nil
}

func ChkFileExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func WorkingDir() string {
	// Working Directory
	wd, _ := os.Getwd()
	return wd
}

func RetryFunc(retrybuilder func() func() error) error {
	ExpBackoff := backoff.NewExponentialBackOff()
	ExpBackoff.InitialInterval = backoffII * time.Second
	ExpBackoff.RandomizationFactor = backoffRF
	ExpBackoff.Multiplier = backoffMult
	ExpBackoff.MaxElapsedTime = backoffMaxETime * time.Second

	err := backoff.Retry(retrybuilder(), ExpBackoff)
	if err != nil {
		return err
	}
	return nil
}

// PrintBuildInfos функция печатает параметны билда
func PrintBuildInfos() {
	// для того что бы понять importpath, собираем бинарь и выполняем команду - go tool nm agent | grep utils
	// -ldflags="-X 'github.com/netzen86/collectmetrics/internal/utils.buildVersion=0.0.1' -X 'github.com/netzen86/collectmetrics/internal/utils.buildDate=$(date +'%d/%m/%y')' -X 'github.com/netzen86/collectmetrics/internal/utils.buildCommit=$(git rev-parse --short HEAD)'"

	fmt.Printf("\nBuild version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n\n", buildCommit)
}

func GetLocalIP(logger zap.SugaredLogger) (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("error when get local ip %w", err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			logger.Errorf("error when clogse gz %w", err)
		}
	}()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP.String(), nil
}
