// Package files - пакет для работы с хранилищем типа файл
package files

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories"
)

type Filestorage struct {
	Filename     string
	FilenameTemp string
}

type Producer struct {
	file     *os.File
	writer   *bufio.Writer
	Filename string
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

// NewFileStorage функция подключения к базе данных, param = строка для подключения к БД
func NewFileStorage(ctx context.Context, param string) (*Filestorage, error) {
	var filestorage Filestorage
	filestorage.Filename = param
	filestorage.FilenameTemp = fmt.Sprintf("%stmp", param)
	return &filestorage, nil
}

func (c *Consumer) ReadMetric(metrics *api.MetricsMap, logger zap.SugaredLogger) error {
	metric := api.Metrics{}
	scanner := c.Scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &metric)
		if err != nil {
			logger.Infof("can't unmarshal string %v", err)
			continue
		}
		switch {
		case metric.MType == api.Gauge:
			if metric.Value == nil {
				return fmt.Errorf(" gauge value is nil %v", err)

			}
			value := float64(*metric.Value)
			metrics.Metrics[metric.ID] = api.Metrics{ID: metric.ID, MType: metric.MType, Value: &value}
		case metric.MType == api.Counter:
			if metric.Delta == nil {
				return fmt.Errorf(" counter delta is nil %v", err)

			}
			delta := int64(*metric.Delta)
			metrics.Metrics[metric.ID] = api.Metrics{ID: metric.ID, MType: metric.MType, Delta: &delta}
		default:
			return fmt.Errorf("rm func - wrong metric type")
		}
		metric.Clean()
	}
	return nil
}

// SaveMetrics функция для сохранения метрик в файл
// использую log.Fatal а не возврат ошибки потому что эта функция будет запускаться в горутине
func SaveMetrics(storage repositories.Repo, metricFileName string,
	storeInterval int, serverCtx context.Context, wg *sync.WaitGroup, logger zap.SugaredLogger) {
	var shutdown bool = false

	for !shutdown {
		<-time.After(time.Duration(storeInterval) * time.Second)
		metrics, err := storage.GetAllMetrics(context.TODO(), logger)
		if err != nil {
			logger.Fatalf("error when getting all metrics %v", err)
		}

		logger.Infoln("ENTER PRODUCER IN SM")
		producer, err := NewProducer(metricFileName)
		if err != nil {
			logger.Fatal("can't create producer")
		}
		for _, metric := range metrics.Metrics {
			logger.Debugf("METRIC %s WRITE IN FILE", metric.MType)
			err = producer.WriteMetric(metric)
			if err != nil {
				logger.Fatal("can't write metric")
			}
		}
		err = producer.file.Close()
		if err != nil {
			logger.Fatal("can't close file")
		}
		select {
		case <-serverCtx.Done():
			logger.Info("stop saving metrics")
			shutdown = true
			wg.Done()
		default:
		}
	}
}

func SyncSaveMetrics(metrics api.MetricsMap, metricFileName string,
	logger zap.SugaredLogger) error {
	producer, err := NewProducer(metricFileName)
	if err != nil {
		return fmt.Errorf("can't create producer %w", err)
	}
	defer func() {
		err = producer.file.Close()
		if err != nil {
			logger.Errorf("can't close file %v", err)
		}
	}()

	err = producer.file.Truncate(0)
	if err != nil {
		return fmt.Errorf("can't create producer %w", err)
	}
	for _, metric := range metrics.Metrics {
		logger.Debugf("METRIC", metric)
		err := producer.WriteMetric(metric)
		if err != nil {
			return fmt.Errorf("can't write metric %v %v %w", metric.ID, metric.MType, err)
		}

	}
	return nil
}
func LoadMetric(metrics *api.MetricsMap, metricFileName string,
	logger zap.SugaredLogger) error {
	if _, err := os.Stat(metricFileName); err == nil {
		consumer, err := NewConsumer(metricFileName)
		if err != nil {
			return fmt.Errorf("can't create consumer in lm %w", err)
		}
		defer func() {
			err = consumer.file.Close()
			if err != nil {
				logger.Errorf("error when closing consumer %v", err)
			}
		}()

		err = consumer.ReadMetric(metrics, logger)
		if err != nil {
			return fmt.Errorf("can't read metric in lm %w", err)
		}
	}
	return nil
}

// функция для накапливания значений в counter
func (fs *Filestorage) sumPc(ctx context.Context, delta int64,
	pcMetric *api.Metrics, logger zap.SugaredLogger) error {

	getdelta, err := fs.GetCounterMetric(ctx, pcMetric.ID, logger)
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

func (fs *Filestorage) UpdateParam(ctx context.Context, cntSummed bool,
	metricType, metricName string,
	metricValue interface{}, logger zap.SugaredLogger) error {
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
			err = fs.sumPc(ctx, delta, &pcMetric, logger)
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
	err = LoadMetric(&metrics, fs.FilenameTemp, logger)
	if err != nil {
		return fmt.Errorf("updateparam error load metrics from file %w", err)
	}

	metrics.Metrics[metricName] = pcMetric

	err = SyncSaveMetrics(metrics, fs.FilenameTemp, logger)
	if err != nil {
		return fmt.Errorf("updateparam error save metirc to file %w", err)

	}
	return nil
}

func (fs *Filestorage) GetCounterMetric(ctx context.Context, metricID string,
	logger zap.SugaredLogger) (int64, error) {
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

func (fs *Filestorage) GetGaugeMetric(ctx context.Context, metricID string,
	logger zap.SugaredLogger) (float64, error) {
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

func (fs *Filestorage) GetAllMetrics(ctx context.Context, logger zap.SugaredLogger) (api.MetricsMap, error) {
	var metrics api.MetricsMap
	metrics.Metrics = make(map[string]api.Metrics)

	if _, err := os.Stat(fs.FilenameTemp); err == nil {
		consumer, err := NewConsumer(fs.FilenameTemp)
		if err != nil {
			return api.MetricsMap{}, fmt.Errorf("can't create consumer in lm %w", err)
		}
		defer func() {
			err = consumer.file.Close()
			if err != nil {
				logger.Errorf("error when closing consumer %v", err)
			}
		}()

		err = consumer.ReadMetric(&metrics, logger)
		if err != nil {
			return api.MetricsMap{}, fmt.Errorf("can't read metric in lm %w", err)
		}
	}
	return metrics, nil
}

func (fs *Filestorage) CreateTables(ctx context.Context, logger zap.SugaredLogger) error {
	return nil
}
