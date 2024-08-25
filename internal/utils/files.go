package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

type Consumer struct {
	file *os.File
	// добавляем Reader в Consumer
	// reader  *bufio.Reader
	scanner *bufio.Scanner
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteMetric(metric api.Metrics) error {
	data, err := json.Marshal(&metric)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый Reader
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadMetric(storage *memstorage.MemStorage) error {
	metric := api.Metrics{}
	ctx := context.Background()
	scanner := c.scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &metric)
		if err != nil {
			log.Printf("can't unmarshal string %v", err)
		}
		if metric.MType == "gauge" {
			err := storage.UpdateParam(ctx, metric.MType, metric.ID, *metric.Value)
			if err != nil {
				return fmt.Errorf("gauge error %v", err)
			}
		}
		if metric.MType == "counter" {
			err := storage.UpdateParam(ctx, metric.MType, metric.ID, *metric.Delta)
			if err != nil {
				return fmt.Errorf("counter error %v", err)
			}
		}
	}
	return nil
}

func SaveMetrics(storage *memstorage.MemStorage, metricFileName string, storeInterval int) {
	for {
		<-time.After(time.Duration(storeInterval) * time.Second)
		log.Println("CREATE PRODUCER")
		producer, err := NewProducer(metricFileName)
		if err != nil {
			log.Fatal("can't create producer")
		}
		for k, v := range storage.Gauge {
			log.Println("METRICS WRITE")
			err := producer.WriteMetric(api.Metrics{MType: "gauge", ID: k, Value: &v})
			if err != nil {
				log.Fatal("can't write metric")
			}
		}
	}
}

func SyncSaveMetrics(storage *memstorage.MemStorage, metricFileName string) {
	producer, err := NewProducer(metricFileName)
	if err != nil {
		log.Fatal("can't create producer")
	}
	for k, v := range storage.Gauge {
		log.Println("METRICS WRITE")
		err := producer.WriteMetric(api.Metrics{MType: "gauge", ID: k, Value: &v})
		if err != nil {
			log.Fatal("can't write metric")
		}
	}
}

func LoadMetric(storage *memstorage.MemStorage, metricFileName string) {
	consumer, err := NewConsumer(metricFileName)
	if err != nil {
		log.Fatal(err, "can't create consumer")
	}
	err = consumer.ReadMetric(storage)
	if err != nil {
		log.Fatal(err, "can't read metric")
	}
}
