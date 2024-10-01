package utils

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/netzen86/collectmetrics/internal/api"
)

const (
	backoffII       time.Duration = 1
	backoffRF       float64       = 0.5
	backoffMult     float64       = 2
	backoffMaxETime time.Duration = 5
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

func GzipDecompress(buf *bytes.Buffer) error {
	// переменная buf будет читать входящие данные и распаковывать их
	log.Println("*** DECOMPRESSING ***")
	gz, err := gzip.NewReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("!!!%s!!! unpacking data error", err)
	}
	defer gz.Close()

	// в отчищенную переменную buf записываются распакованные данные
	buf.Reset()
	// _, err =
	buf.ReadFrom(gz)
	// if err != nil {
	// 	return fmt.Errorf("%s read unpacked data error", err)
	// }
	return nil
}

func SelectDeCoHTTP(buf *bytes.Buffer, r interface{}) error {
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
		err := GzipDecompress(buf)
		if err != nil {
			return fmt.Errorf("!!%s!! unpack data error", err)
		}
	}
	return nil
}

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

func ListDir(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return nil
}

func RetrayFunc(retrybuilder func() func() error) error {
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
