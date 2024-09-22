package files

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
)

type Producer struct {
	file     *os.File
	Filename string
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
		file:     file,
		Filename: filename,
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
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
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
			if metric.Delta == nil {
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

func SaveMetrics(storage *memstorage.MemStorage, metricFileName, tempfile, storageSelecter string, storeInterval int) {

	for {
		<-time.After(time.Duration(storeInterval) * time.Second)
		if storageSelecter == "FILE" {
			log.Println("LOAD METRIC FORM TEMP FILE")
			LoadMetric(storage, tempfile)
		}
		log.Println(storage.Counter, storage.Gauge, tempfile)

		log.Println("ENTER PRODUCER IN SM")
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
		producer.file.Close()
	}
}

func SyncSaveMetrics(storage *memstorage.MemStorage, metricFileName string) {
	producer, err := NewProducer(metricFileName)
	if err != nil {
		log.Fatal("can't create producer")
	}
	defer producer.file.Close()
	for k, v := range storage.Gauge {
		err := producer.WriteMetric(api.Metrics{MType: "gauge", ID: k, Value: &v})
		if err != nil {
			log.Fatal("can't write metric")
		}
	}
	for k, v := range storage.Counter {
		err := producer.WriteMetric(api.Metrics{MType: "counter", ID: k, Delta: &v})
		if err != nil {
			log.Fatal("can't write metric")
		}
	}

}

func LoadMetric(storage *memstorage.MemStorage, metricFileName string) {
	if _, err := os.Stat(metricFileName); err == nil {
		consumer, err := NewConsumer(metricFileName)
		if err != nil {
			log.Fatal(err, " can't create consumer in lm")
		}
		defer consumer.file.Close()
		err = consumer.ReadMetric(storage)
		if err != nil {
			log.Fatal(err, " can't read metric in lm")
		}
	}
}

func UpdateParamFile(ctx context.Context, saveMetricsDefaultPath, metricType, metricName string, metricValue interface{}) error {
	producer, err := NewProducer(saveMetricsDefaultPath)
	if err != nil {
		log.Fatal("can't create producer")
	}
	defer producer.file.Close()
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

func ReadOneMetric(ctx context.Context, consumer *Consumer, metric *api.Metrics) (string, error) {
	var scannedMetric api.Metrics
	scanner := consumer.Scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &scannedMetric)
		if err != nil {
			log.Printf(" can't unmarshal string %v", err)
		}
		if scannedMetric.ID == metric.ID {
			switch {
			case scannedMetric.MType == "gauge":
				if scannedMetric.Value != nil {
					metric.Value = scannedMetric.Value
					return "", nil
				} else {
					return "", fmt.Errorf(" gauge vlaue is nil %v", err)
				}
			case metric.MType == "counter":

				if scannedMetric.Delta != nil {
					metric.Delta = scannedMetric.Delta
					return "", nil
				} else {
					return "", fmt.Errorf(" counter delta is nil %v", err)
				}
			default:
				return fmt.Sprintf(" metric %s %s wrong type ", metric.ID, metric.MType), nil
			}
		}
	}
	return "", fmt.Errorf("metric %s %s not exist ", metric.ID, metric.MType)
}

func sumPc(ctx context.Context, filename string, delta int64, pcMetric *api.Metrics) error {
	consumer, err := NewConsumer(filename)
	if err != nil {
		return err
	}
	_, err = ReadOneMetric(ctx, consumer, pcMetric)
	if err != nil {
		if strings.Contains(err.Error(), "not exist") {
			*pcMetric.Delta = 0
		} else {
			return err
		}
	}

	*pcMetric.Delta = *pcMetric.Delta + delta

	return nil
}

func FileStorage(ctx context.Context, tempfile string, metric api.Metrics) error {
	var pcMetric api.Metrics
	var err error

	tmpStorage, err := memstorage.NewMemStorage()
	if err != nil {
		return err
	}
	pcMetric.ID = metric.ID
	pcMetric.MType = metric.MType

	if metric.MType == "counter" {
		pcMetric.Delta = metric.Delta
		err = sumPc(ctx, tempfile, *metric.Delta, &pcMetric)
		if err != nil {
			return err
		}
	} else if metric.MType == "gauge" {
		pcMetric.Value = metric.Value
	}

	LoadMetric(tmpStorage, tempfile)

	if metric.MType == "counter" {
		tmpStorage.Counter[metric.ID] = *pcMetric.Delta
	} else if metric.MType == "gauge" {
		tmpStorage.Gauge[metric.ID] = *metric.Value
	}

	SyncSaveMetrics(tmpStorage, tempfile)
	if err != nil {
		return err
	}
	return nil
}
