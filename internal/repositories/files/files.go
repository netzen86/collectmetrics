package files

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
	"github.com/netzen86/collectmetrics/internal/utils"
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
	Scanner *bufio.Scanner
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
		Scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadMetric(storage *memstorage.MemStorage) error {
	metric := api.Metrics{}
	ctx := context.Background()
	scanner := c.Scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &metric)
		if err != nil {
			log.Printf("can't unmarshal string %v", err)
		}
		if metric.MType == "gauge" {
			if metric.Value == nil {
				return fmt.Errorf(" gauge value is nil %v", err)

			}
			err := storage.UpdateParam(ctx, true, metric.MType, metric.ID, *metric.Value)
			if err != nil {
				return fmt.Errorf("gauge error %v", err)
			}
		}
		if metric.MType == "counter" {
			if metric.Value == nil {
				return fmt.Errorf(" counter delta is nil %v", err)

			}
			err := storage.UpdateParam(ctx, true, metric.MType, metric.ID, *metric.Delta)
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
			log.Println("METRICS GAUGE WRITE")
			err := producer.WriteMetric(api.Metrics{MType: "gauge", ID: k, Value: &v})
			if err != nil {
				log.Fatal("can't write metric")
			}
		}
		for k, v := range storage.Counter {
			log.Println("METRICS COUNTER WRITE")
			err := producer.WriteMetric(api.Metrics{MType: "counter", ID: k, Delta: &v})
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
	for k, v := range storage.Counter {
		log.Println("METRICS COUNTER WRITE")
		err := producer.WriteMetric(api.Metrics{MType: "counter", ID: k, Delta: &v})
		if err != nil {
			log.Fatal("can't write metric")
		}
	}
}

func LoadMetric(storage *memstorage.MemStorage, metricFileName string) {
	consumer, err := NewConsumer(metricFileName)
	if err != nil {
		log.Fatal(err, " can't create consumer in lm")
	}
	err = consumer.ReadMetric(storage)
	if err != nil {
		log.Fatal(err, " can't read metric in lm")
	}
}

func UpdateParamFile(ctx context.Context, producer *Producer, metricType, metricName string, metricValue interface{}) error {

	if metricType == "gauge" {
		// log.Println("METRICS GAUGE WRITE TO FILE")
		val, err := utils.ParseValGag(metricValue)
		if err != nil {
			return fmt.Errorf("gauge value error %v", err)
		}
		err = producer.WriteMetric(api.Metrics{MType: "gauge", ID: metricName, Value: &val})
		if err != nil {
			return fmt.Errorf("can't write gauge metric %v", err)
		}
	} else if metricType == "counter" {
		// log.Println("METRICS COUNTER WRITE TO FILE")
		del, err := utils.ParseValCnt(metricValue)
		if err != nil {
			return fmt.Errorf("counter value error %v", err)
		}
		err = producer.WriteMetric(api.Metrics{MType: "counter", ID: metricName, Delta: &del})
		if err != nil {
			return fmt.Errorf("can't write counter metric %v", err)
		}
	} else {
		return fmt.Errorf("%s", "wrong metric type")
	}
	return nil
}

func ReadOneMetric(ctx context.Context, consumer *Consumer, metric *api.Metrics) error {
	var scannedMetric api.Metrics
	scanner := consumer.Scanner
	log.Println("!!!! enter one metric read")
	if len(scanner.Bytes()) == 0 {
		return fmt.Errorf(" scanner equal %s", "nil")
	}
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &scannedMetric)
		if err != nil {
			return fmt.Errorf(" can't unmarshal string %v", err)
		}
		if scannedMetric.ID == metric.ID {
			if metric.MType == "gauge" {
				if scannedMetric.Value != nil {
					metric.Value = scannedMetric.Value
				} else {
					return fmt.Errorf(" gauge vlaue is nil %v", err)
				}
			}
			if metric.MType == "counter" {

				if scannedMetric.Delta != nil {
					metric.Delta = scannedMetric.Delta
				} else {
					return fmt.Errorf(" counter delta is nil %v", err)
				}
			}
		} else {
			return fmt.Errorf(" metric %s %s not exist ", metric.ID, metric.MType)
		}
	}
	return nil
}
