package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
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

func GzipDecompress(data []byte) ([]byte, error) {

	// переменная r будет читать входящие данные и распаковывать их
	gz, err := gzip.NewReader(bytes.NewReader(data))
	defer gz.Close()

	var b bytes.Buffer
	// в переменную b записываются распакованные данные
	_, err = b.ReadFrom(gz)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
