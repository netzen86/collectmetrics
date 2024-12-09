// Package files - пакет для работы с хранилищем типа файл
package files

import (
	"bufio"
	"os"
	"testing"

	"github.com/netzen86/collectmetrics/internal/api"
	"go.uber.org/zap"
)

func TestProducer_WriteMetric(t *testing.T) {
	type fields struct {
		file     *os.File
		writer   *bufio.Writer
		Filename string
	}
	type args struct {
		metric api.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Producer{
				file:     tt.fields.file,
				writer:   tt.fields.writer,
				Filename: tt.fields.Filename,
			}
			if err := p.WriteMetric(tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("Producer.WriteMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConsumer_ReadMetric(t *testing.T) {
	type fields struct {
		file    *os.File
		Scanner *bufio.Scanner
	}
	type args struct {
		metrics *api.MetricsMap
		logger  zap.SugaredLogger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Consumer{
				file:    tt.fields.file,
				Scanner: tt.fields.Scanner,
			}
			if err := c.ReadMetric(tt.args.metrics, tt.args.logger); (err != nil) != tt.wantErr {
				t.Errorf("Consumer.ReadMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
