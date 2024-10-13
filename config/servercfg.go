package config

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/db"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
)

const (
	addressServer    string = "localhost:8080"
	storeIntervalDef int    = 300
	// имена переменных окружения
	envAdd string = "ADDRESS"
	envSI  string = "STORE_INTERVAL"
	envFSP string = "FILE_STORAGE_PATH"
	envRes string = "RESTORE"
	envKey string = "KEY"
	envDB  string = "DATABASE_DSN"
)

type ServerCfg struct {
	// адрес и порт на котором запуститься сервер
	Endpoint string `env:"ADDRESS" DefVal:"localhost:8080"`
	// имя и путь к файлу для хранения метрик
	FileStoragePath string `env:"FILE_STORAGE_PATH" DefVal:""`
	// имя и путь к файлу для хранения метрик для значения по умолчанию
	FileStoragePathDef string `env:"" DefVal:"FileStoragePath"`
	// ключ для создания подписи данных
	SignKeyString string `env:"KEY" DefVal:""`
	// строка для подключения к базе данных
	DBconstring string `env:"DATABASE_DSN" DefVal:""`
	// ключ для выбора текущего хранилища (мемстораж, файл, база данных)
	StorageSelecter string `env:"" DefVal:"MEMORY"`
	// интервал сохранения метрик в файл
	StoreInterval int `env:"STORE_INTERVAL" DefVal:"300s"`
	// ключ для определения восстановления метрик из файла
	Restore bool `env:"RESTORE" DefVal:"true"`
	// указатель на временный файл хранения метрик
	Tempfile *os.File `env:"" DefVal:""`
	// указатель на memstorage
	Storage repositories.Repo `env:"" DefVal:""`
}

// метод для получения параметров запуска сервера из флагов
func (serverCfg *ServerCfg) parseSrvFlags() error {

	// имя файла для сохранения метрик по умолчанию
	serverCfg.FileStoragePathDef = "servermetrics.json"

	// опредаляем флаги
	flag.StringVar(&serverCfg.Endpoint, "a", addressServer, "Used to set the address and port on which the server runs.")
	flag.StringVar(&serverCfg.FileStoragePath, "f", serverCfg.FileStoragePathDef, "Used to set file path to save metrics.")
	flag.StringVar(&serverCfg.DBconstring, "d", "", "Used to set db connet string.")
	flag.StringVar(&serverCfg.SignKeyString, "k", "", "Used to set key for calc hash.")
	flag.BoolVar(&serverCfg.Restore, "r", true, "Used to set restore metrics.")
	flag.IntVar(&serverCfg.StoreInterval, "i", storeIntervalDef, "Used for set save metrics on disk.")

	flag.Parse()

	// если серверу преданы параменты, а не флаги
	// печатаем какие параметры можно использовать
	if len(flag.Args()) != 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// для замены имени файла по умолчанию
	if serverCfg.FileStoragePath != serverCfg.FileStoragePathDef {
		serverCfg.FileStoragePathDef = serverCfg.FileStoragePath
	}
	return nil
}

// метод для получения параметров запуска сервера из переменных окружения
func (serverCfg *ServerCfg) getSrvEnv() error {
	var err error

	// получаем данные для работы програмы из переменных окружения
	// переменные окружения имеют наивысший приоритет
	if len(os.Getenv(envAdd)) > 0 {
		serverCfg.Endpoint = os.Getenv(envAdd)
	}
	// получаем интервал сохранения метрик в файл
	if len(os.Getenv(envSI)) != 0 {
		serverCfg.StoreInterval, err = strconv.Atoi(os.Getenv(envSI))
		if err != nil {
			return fmt.Errorf("error atoi poll interval %v ", err)
		}
	}

	// получаем имя файла для сохранения метрик
	if len(os.Getenv(envFSP)) != 0 {
		serverCfg.FileStoragePath = os.Getenv(envFSP)
		serverCfg.FileStoragePathDef = os.Getenv(envFSP)

	}
	// получаем параметр восстановления метрик из файла
	// при false не восстанавливаем, по умолчанию true
	if len(os.Getenv(envRes)) != 0 {
		serverCfg.Restore, err = strconv.ParseBool(os.Getenv(envRes))
		if err != nil {
			return fmt.Errorf("error parse bool restore %v ", err)
		}
	}
	// получаем ключ для создания подписи
	if len(os.Getenv(envKey)) != 0 {
		serverCfg.SignKeyString = os.Getenv(envKey)
	}
	// получаем параметры подключения к базе данных
	if len(os.Getenv(envDB)) != 0 {
		serverCfg.DBconstring = os.Getenv(envDB)
	}
	return nil
}

// метод инициализации сервера
func (serverCfg *ServerCfg) initSrv() error {
	var err error
	ctx := context.Background()

	// созданиe мемсторэжа
	if serverCfg.FileStoragePath == serverCfg.FileStoragePathDef &&
		len(serverCfg.DBconstring) == 0 {
		serverCfg.Storage = memstorage.NewMemStorage()
	}

	// созданиe базы данных
	if len(serverCfg.DBconstring) != 0 {
		serverCfg.Storage, err = db.NewDBStorage(ctx, serverCfg.DBconstring)
		if err != nil {
			return fmt.Errorf("error when get mem storage %v ", err)
		}
		// создание таблиц counter и gauge
		retrybuilder := func() func() error {
			return func() error {
				err := serverCfg.Storage.CreateTables(ctx)
				if err != nil {
					log.Println(err)
				}
				return err
			}
		}
		err = utils.RetrayFunc(retrybuilder)
		if err != nil {
			log.Fatal("tables not created ", err)
		}
	}

	// созданиe файлсторэжа
	if serverCfg.FileStoragePath != serverCfg.FileStoragePathDef {
		serverCfg.Storage, err = files.NewFileStorage(ctx, serverCfg.FileStoragePath)
		if err != nil {
			return fmt.Errorf("error when get file storage %v ", err)
		}
		// создание временого файла для файлсторож
		serverCfg.Tempfile, err = os.OpenFile(fmt.Sprintf("%stmp",
			serverCfg.FileStoragePath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return fmt.Errorf("error create temp file %v ", err)
		}
	}

	// лог значений полученных из переменных окружения и флагов
	log.Println("!!! SERVER CONFIGURED !!!",
		serverCfg.Endpoint, serverCfg.FileStoragePathDef,
		serverCfg.FileStoragePath, serverCfg.DBconstring,
		len(serverCfg.SignKeyString), serverCfg.Restore, serverCfg.StoreInterval)

	// копируем метрики из файла в мемсторож
	if serverCfg.Restore {
		var metrics api.MetricsMap
		metrics.Metrics = make(map[string]api.Metrics)
		log.Println("ENTER IN RESTORE")

		err = files.LoadMetric(&metrics, serverCfg.FileStoragePathDef)
		if err != nil {
			return fmt.Errorf("error load metrics fom file %w", err)
		}
		for _, metric := range metrics.Metrics {
			if metric.MType == api.Gauge {
				err := serverCfg.Storage.UpdateParam(ctx, false, metric.MType, metric.ID, *metric.Value)
				if err != nil {
					return fmt.Errorf("error restore lm %s %s : %w", metric.ID, metric.MType, err)
				}
			} else if metric.MType == api.Counter {
				err := serverCfg.Storage.UpdateParam(ctx, false, metric.MType, metric.ID, *metric.Delta)
				if err != nil {
					return fmt.Errorf("error restore lm %s %s : %w", metric.ID, metric.MType, err)
				}
			}
		}
	}
	return nil
}

// метод для получения конфигурации сервера
func (serverCfg *ServerCfg) GetServerCfg() error {

	err := serverCfg.parseSrvFlags()
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	err = serverCfg.getSrvEnv()
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	err = serverCfg.initSrv()
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	return nil
}
