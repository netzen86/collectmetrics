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
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

type filestorage struct {
	Filename     string
	FilenameTemp string
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

// функция подключения к базе данных, param = строка для подключения к БД
func NewFileStorage(ctx context.Context, param string) (*filestorage, error) {
	var filestorage filestorage
	filestorage.Filename = param
	filestorage.FilenameTemp = fmt.Sprintf("%stmp", param)
	return &filestorage, nil
}

func (c *Consumer) ReadMetric(metrics *api.MetricsMap) error {
	metric := api.Metrics{}
	scanner := c.Scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &metric)
		if err != nil {
			log.Printf("can't unmarshal string %v", err)
			continue
		}
		if metric.MType == "gauge" {
			if metric.Value == nil {
				return fmt.Errorf(" gauge value is nil %v", err)

			}
			value := float64(*metric.Value)
			metrics.Metrics[metric.ID] = api.Metrics{ID: metric.ID, MType: metric.MType, Value: &value}
		} else if metric.MType == "counter" {
			if metric.Delta == nil {
				return fmt.Errorf(" counter delta is nil %v", err)

			}
			delta := int64(*metric.Delta)
			metrics.Metrics[metric.ID] = api.Metrics{ID: metric.ID, MType: metric.MType, Delta: &delta}
		} else {
			return fmt.Errorf("rm func - wrong metric type")
		}
		metric.Clean()
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

func SyncSaveMetrics(metrics api.MetricsMap, metricFileName string) error {
	producer, err := NewProducer(metricFileName)
	if err != nil {
		return fmt.Errorf("can't create producer %w", err)
	}
	defer producer.file.Close()
	err = producer.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("can't create producer %w", err)
	}
	for _, metric := range metrics.Metrics {
		// log.Panicln("METRIC", metric)
		err := producer.WriteMetric(metric)
		if err != nil {
			return fmt.Errorf("can't write metric %v %v %w", metric.ID, metric.MType, err)
		}

	}
	return nil
}
func LoadMetric(metrics *api.MetricsMap, metricFileName string) error {
	if _, err := os.Stat(metricFileName); err == nil {
		consumer, err := NewConsumer(metricFileName)
		if err != nil {
			return fmt.Errorf("can't create consumer in lm %w", err)
		}
		defer consumer.file.Close()
		err = consumer.ReadMetric(metrics)
		if err != nil {
			return fmt.Errorf("can't read metric in lm %w", err)
		}
	}
	return nil
}

// функция для накапливания значений в counter
func (fs *filestorage) sumPc(ctx context.Context, delta int64, pcMetric *api.Metrics) error {

	getdelta, err := fs.GetCounterMetric(ctx, pcMetric.ID)
	pcMetric.Delta = &getdelta
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

func (fs *filestorage) UpdateParam(ctx context.Context, cntSummed bool,
	metricType, metricName string, metricValue interface{}) error {
	var pcMetric api.Metrics
	var metrics api.MetricsMap
	var err error

	metrics.Metrics = make(map[string]api.Metrics)
	pcMetric.ID = metricName
	pcMetric.MType = metricType

	// в зависимости от типа метрик определяем тип metricValue
	if metricType == api.Counter {
		delta, ok := metricValue.(int64)
		if !ok {
			return fmt.Errorf("mismatch metric %s and value type in filestorage", metricName)
		}
		pcMetric.Delta = &delta
		if cntSummed {
			err = fs.sumPc(ctx, delta, &pcMetric)
			if err != nil {
				return err
			}
		}
	} else if metricType == api.Gauge {
		value, ok := metricValue.(float64)
		if !ok {
			return fmt.Errorf("mismatch metric %s and value type in filestorage", metricName)
		}
		pcMetric.Value = &value
	}
	err = LoadMetric(&metrics, fs.FilenameTemp)
	if err != nil {
		return fmt.Errorf("updateparam error load metrics from file %w", err)
	}

	metrics.Metrics[metricName] = pcMetric

	err = SyncSaveMetrics(metrics, fs.FilenameTemp)
	if err != nil {
		return fmt.Errorf("updateparam error save metirc to file %w", err)

	}

	return nil
}

func (fs *filestorage) GetCounterMetric(ctx context.Context, metricID string) (int64, error) {
	var scannedMetric api.Metrics
	consumer, err := NewConsumer(fs.FilenameTemp)
	if err != nil {
		return 0, fmt.Errorf("can't create consumer %w", err)
	}
	scanner := consumer.Scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &scannedMetric)
		if err != nil {
			return 0, fmt.Errorf("can't unmarshal string %w", err)
		}
		if scannedMetric.ID == metricID {
			if scannedMetric.Delta != nil {
				return *scannedMetric.Delta, nil
			} else {
				return 0, fmt.Errorf("counter delta is nil %w", err)
			}
		}
	}
	return 0, fmt.Errorf("metric %s %s not exist ", metricID, api.Counter)
}

func (fs *filestorage) GetGaugeMetric(ctx context.Context, metricID string) (float64, error) {
	var scannedMetric api.Metrics
	consumer, err := NewConsumer(fs.FilenameTemp)
	if err != nil {
		return 0, fmt.Errorf("can't create consumer %w", err)
	}
	scanner := consumer.Scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &scannedMetric)
		if err != nil {
			return 0, fmt.Errorf("can't unmarshal string %w", err)
		}
		if scannedMetric.ID == metricID {
			if scannedMetric.Value != nil {
				return *scannedMetric.Value, nil
			} else {
				return 0, fmt.Errorf("gauge value is nil %w", err)
			}
		}
	}
	return 0, fmt.Errorf("metric %s %s not exist ", metricID, api.Gauge)
}

func (fs *filestorage) GetAllMetrics(ctx context.Context) (api.MetricsMap, error) {
	var metrics api.MetricsMap
	metrics.Metrics = make(map[string]api.Metrics)

	if _, err := os.Stat(fs.FilenameTemp); err == nil {
		consumer, err := NewConsumer(fs.FilenameTemp)
		if err != nil {
			return api.MetricsMap{}, fmt.Errorf("can't create consumer in lm %w", err)
		}
		defer consumer.file.Close()
		err = consumer.ReadMetric(&metrics)
		if err != nil {
			return api.MetricsMap{}, fmt.Errorf("can't read metric in lm %w", err)
		}
	}
	return metrics, nil
}

func (fs *filestorage) GetStorage(ctx context.Context) (*memstorage.MemStorage, error) {
	return &memstorage.MemStorage{}, nil
}

func (fs *filestorage) CreateTables(ctx context.Context) error {
	return nil
}
