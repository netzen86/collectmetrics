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
	"github.com/netzen86/collectmetrics/internal/repositories"
)

type filestorage struct {
	Producer
	Consumer
}

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

func (c *Consumer) ReadMetric(metrics *api.MetricsMap) error {
	metric := api.Metrics{}
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
		} else if metric.MType == "counter" {
			if metric.Delta == nil {
				return fmt.Errorf(" counter delta is nil %v", err)

			}
		} else {
			return fmt.Errorf("rm func - wrong metric type")

		}
		metrics.Metrics[metric.ID] = metric
	}
	return nil
}

// функция для сохранения метрик в файл
// использую log.Fatal а не возврат ошибки потому что эта функция будет запускаться в горутине
func SaveMetrics(storage repositories.Repo, metricFileName, tempfile, storageSelecter string, storeInterval int) {

	for {
		<-time.After(time.Duration(storeInterval) * time.Second)
		metrics, err := storage.GetAllMetrics(context.TODO())
		if err != nil {
			log.Fatalf("error when getting all metrics %v", err)
		}

		log.Println("ENTER PRODUCER IN SM")
		producer, err := NewProducer(metricFileName)
		if err != nil {
			log.Fatal("can't create producer")
		}
		for _, metric := range metrics.Metrics {
			log.Printf("METRIC %s WRITE IN FILE", metric.MType)
			err := producer.WriteMetric(metric)
			if err != nil {
				log.Fatal("can't write metric")
			}
		}
		producer.file.Close()
	}
}

func SyncSaveMetrics(metrics api.MetricsMap, metricFileName string) {
	producer, err := NewProducer(metricFileName)
	if err != nil {
		log.Fatal("can't create producer")
	}
	defer producer.file.Close()
	for _, metric := range metrics.Metrics {
		err := producer.WriteMetric(metric)
		if err != nil {
			log.Fatal("can't write metric")
		}

	}
}
func LoadMetric(metrics *api.MetricsMap, metricFileName string) {
	if _, err := os.Stat(metricFileName); err == nil {
		consumer, err := NewConsumer(metricFileName)
		if err != nil {
			log.Fatal(err, " can't create consumer in lm")
		}
		defer consumer.file.Close()
		err = consumer.ReadMetric(metrics)
		if err != nil {
			log.Fatal(err, " can't read metric in lm")
		}
	}
}

// func UpdateParamFile(ctx context.Context, saveMetricsDefaultPath, metricType, metricName string, metricValue interface{}) error {
// 	producer, err := NewProducer(saveMetricsDefaultPath)
// 	if err != nil {
// 		log.Fatal("can't create producer")
// 	}
// 	defer producer.file.Close()
// 	if metricType == "gauge" {
// 		// log.Println("METRICS GAUGE WRITE TO FILE")
// 		val, err := utils.ParseValGag(metricValue)
// 		if err != nil {
// 			return fmt.Errorf("gauge value error %v", err)
// 		}
// 		err = producer.WriteMetric(api.Metrics{MType: "gauge", ID: metricName, Value: &val})
// 		if err != nil {
// 			return fmt.Errorf("can't write gauge metric %v", err)
// 		}
// 	} else if metricType == "counter" {
// 		// log.Println("METRICS COUNTER WRITE TO FILE")
// 		del, err := utils.ParseValCnt(metricValue)
// 		if err != nil {
// 			return fmt.Errorf("counter value error %v", err)
// 		}
// 		err = producer.WriteMetric(api.Metrics{MType: "counter", ID: metricName, Delta: &del})
// 		if err != nil {
// 			return fmt.Errorf("can't write counter metric %v", err)
// 		}
// 	} else {
// 		return fmt.Errorf("%s", "wrong metric type")
// 	}
// 	return nil
// }

func ReadOneMetric(ctx context.Context, consumer *Consumer, metric *api.Metrics) (string, error) {
	var scannedMetric api.Metrics
	log.Println("consumer", consumer.file.Name())
	scanner := consumer.Scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &scannedMetric)
		if err != nil {
			return "", fmt.Errorf(" can't unmarshal string %w", err)
		}
		if scannedMetric.ID == metric.ID {
			switch {
			case scannedMetric.MType == "gauge":
				if scannedMetric.Value != nil {
					metric.Value = scannedMetric.Value
					return "", nil
				} else {
					return "", fmt.Errorf(" gauge vlaue is nil %w", err)
				}
			case metric.MType == "counter":

				if scannedMetric.Delta != nil {
					metric.Delta = scannedMetric.Delta
					return "", nil
				} else {
					return "", fmt.Errorf(" counter delta is nil %w", err)
				}
			default:
				return fmt.Sprintf(" metric %s %s wrong type ", metric.ID, metric.MType), nil
			}
		}
	}
	return "", fmt.Errorf("metric %s %s not exist ", metric.ID, metric.MType)
}

// функция для накапливания значений в counter
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

// func FileStorage(ctx context.Context, tempfile string, metric api.Metrics) error {
// 	var pcMetric api.Metrics
// 	var err error

// 	tmpStorage, err := memstorage.NewMemStorage()
// 	if err != nil {
// 		return err
// 	}
// 	pcMetric.ID = metric.ID
// 	pcMetric.MType = metric.MType

// 	if metric.MType == "counter" {
// 		pcMetric.Delta = metric.Delta
// 		err = sumPc(ctx, tempfile, *metric.Delta, &pcMetric)
// 		if err != nil {
// 			return err
// 		}
// 	} else if metric.MType == "gauge" {
// 		pcMetric.Value = metric.Value
// 	}

// 	LoadMetric(tmpStorage, tempfile)

// 	if metric.MType == "counter" {
// 		tmpStorage.Counter[metric.ID] = *pcMetric.Delta
// 	} else if metric.MType == "gauge" {
// 		tmpStorage.Gauge[metric.ID] = *metric.Value
// 	}

// 	SyncSaveMetrics(tmpStorage, tempfile)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (fs *filestorage) UpdateParam(ctx context.Context, cntSummed bool,
	metricType, metricName string, metricValue interface{}) error {
	var pcMetric api.Metrics
	var metrics api.MetricsMap
	var err error

	// tmpStorage, err := memstorage.NewMemStorage()
	// if err != nil {
	// 	return err
	// }
	pcMetric.ID = metricName
	pcMetric.MType = metricType

	// в зависимости от типа метрик определяем тип metricValue
	if metricType == api.Counter {
		delta, ok := metricValue.(int64)
		if !ok {
			return fmt.Errorf("mismatch metric %s and value type in filestorage", metricName)
		}
		// pcMetric.Delta = &delta
		err = sumPc(ctx, fmt.Sprintf("%stmp", fs.Filename), delta, &pcMetric)
		if err != nil {
			return err
		}
	} else if metricType == api.Gauge {
		value, ok := metricValue.(float64)
		if !ok {
			return fmt.Errorf("mismatch metric %s and value type in filestorage", metricName)
		}
		pcMetric.Value = &value
	}

	LoadMetric(&metrics, fmt.Sprintf("%stmp", fs.Filename))

	metrics.Metrics[metricName] = pcMetric

	SyncSaveMetrics(metrics, fmt.Sprintf("%stmp", fs.Filename))
	if err != nil {
		return err
	}
	return nil
}

// func (fs *filestorage) GetCounterMetric(ctx context.Context, metricID string) (int64, error) {
// 	return 0, nil
// }
// func (fs *filestorage) GetGaugeMetric(ctx context.Context, metricID string) (float64, error) {
// 	return 0, nil
// }
// func (fs *filestorage) GetAllMetrics(ctx context.Context) (api.MetricsSlice, error) {
// 	return api.MetricsSlice{}, nil
// }
// func (fs *filestorage) GetStorage(ctx context.Context) (*memstorage.MemStorage, error) {
// 	return &memstorage.MemStorage{}, nil
// }
// func (fs *filestorage) CreateTables(ctx context.Context) error {
// 	return nil
// }
