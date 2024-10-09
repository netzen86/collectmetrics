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
	// значения store selector
	ssFile     string = "FILE"
	ssDataBase string = "DATABASE"
	ssMemStor  string = "MEMORY"
	// имена переменных окружения
	EnvAdd string = "ADDRESS"
	EnvSI  string = "STORE_INTERVAL"
	EnvFSP string = "FILE_STORAGE_PATH"
	EnvRes string = "RESTORE"
	EnvKey string = "KEY"
	EnvDB  string = "DATABASE_DSN"
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

	// значение переменных по умолчанию
	serverCfg.StorageSelecter = ssMemStor
	serverCfg.FileStoragePathDef = "servermetrics.json"

	// опредаляем флаги
	flag.StringVar(&serverCfg.Endpoint, "a", addressServer, "Used to set the address and port on which the server runs.")
	flag.StringVar(&serverCfg.FileStoragePath, "f", serverCfg.FileStoragePathDef, "Used to set file path to save metrics.")
	flag.StringVar(&serverCfg.DBconstring, "d", "", "Used to set db connet string.")
	flag.StringVar(&serverCfg.SignKeyString, "k", "", "Used to set key for calc hash.")
	flag.BoolVar(&serverCfg.Restore, "r", true, "Used to set restore metrics.")
	flag.IntVar(&serverCfg.StoreInterval, "i", storeIntervalDef, "Used for set save metrics on disk.")

	flag.Parse()

	// если серверу преданы параменты а не флаги печатаем какие параметры можно использовать
	if len(flag.Args()) != 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// для земены значения по умолчания имени файла сохранения метрик
	if serverCfg.FileStoragePath != serverCfg.FileStoragePathDef && len(serverCfg.FileStoragePath) != 0 {
		serverCfg.FileStoragePathDef = serverCfg.FileStoragePath
		serverCfg.StorageSelecter = ssFile
	}

	if len(serverCfg.DBconstring) != 0 {
		serverCfg.StorageSelecter = ssDataBase
	}

	return nil
}

// метод для получения параметров запуска сервера из переменных окружения
func (serverCfg *ServerCfg) getSrvEnv() error {
	var err error

	// получаем данные для работы програмы из переменных окружения
	// переменные окружения имеют наивысший приоритет
	if len(os.Getenv(EnvAdd)) > 0 {
		serverCfg.Endpoint = os.Getenv(EnvAdd)
	}

	if len(os.Getenv(EnvSI)) != 0 {
		serverCfg.StoreInterval, err = strconv.Atoi(os.Getenv(EnvSI))
		if err != nil {
			return fmt.Errorf("error atoi poll interval %v ", err)
		}
	}

	if len(os.Getenv(EnvFSP)) != 0 {
		serverCfg.FileStoragePath = os.Getenv(EnvFSP)
		serverCfg.FileStoragePathDef = os.Getenv(EnvFSP)
		serverCfg.StorageSelecter = ssFile

	}

	if len(os.Getenv(EnvRes)) != 0 {
		serverCfg.Restore, err = strconv.ParseBool(os.Getenv(EnvRes))
		if err != nil {
			return fmt.Errorf("error parse bool restore %v ", err)
		}
	}

	if len(os.Getenv(EnvKey)) != 0 {
		serverCfg.SignKeyString = os.Getenv(EnvKey)
	}

	if len(os.Getenv(EnvDB)) != 0 {
		serverCfg.DBconstring = os.Getenv(EnvDB)
		serverCfg.StorageSelecter = ssDataBase
	}
	return nil
}

// метод инициализации сервера
func (serverCfg *ServerCfg) initSrv() error {
	var ctx context.Context
	// var storage repositories.Repo
	var err error

	// serverCfg.Storage = &storage

	// создания мемсторожа
	serverCfg.Storage, err = memstorage.NewMemStorage()
	if err != nil {
		return fmt.Errorf("error when get mem storage %v ", err)
	}
	if serverCfg.StorageSelecter == ssDataBase {
		// создания базы данных
		serverCfg.Storage, err = db.NewDBStorage(ctx, serverCfg.DBconstring)
		if err != nil {
			return fmt.Errorf("error when get mem storage %v ", err)
		}
	}

	// лог значений полученных из переменных окружения и флагов
	log.Println("!!! SERVER START !!!",
		serverCfg.Endpoint, serverCfg.FileStoragePathDef,
		serverCfg.FileStoragePath, serverCfg.DBconstring,
		len(serverCfg.SignKeyString), serverCfg.Restore, serverCfg.StoreInterval)

	// создание временого файла
	serverCfg.Tempfile, err = os.OpenFile(fmt.Sprintf("%stmp", serverCfg.FileStoragePath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("error create temp file %v ", err)
	}

	// если файл с метриками существует копируем его в файлсторож
	if utils.ChkFileExist(serverCfg.FileStoragePath) {
		data, err := os.ReadFile(serverCfg.FileStoragePath) //read the contents of source file
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}
		err = os.WriteFile(fmt.Sprintf("%stmp", serverCfg.FileStoragePath), data, 0666) //write the content to destination file
		if err != nil {
			return fmt.Errorf("error writing file: %v", err)
		}
	}

	// копируем метрики из файла в мемсторож
	if serverCfg.Restore && serverCfg.StorageSelecter == ssMemStor {
		var metrics api.MetricsMap
		log.Println("ENTER IN RESTORE")
		memstor, err := repositories.Repo.GetStorage(serverCfg.Storage, ctx)
		if err != nil {
			return fmt.Errorf("error get storage %w", err)
		}
		files.LoadMetric(&metrics, serverCfg.FileStoragePathDef)
		memstorage.MetricMapToMemstorage(&metrics, *memstor)
	}

	// если храним метрики в базе данных то создаем таблицы counter и gauge
	if serverCfg.StorageSelecter == ssDataBase {
		retrybuilder := func() func() error {
			return func() error {
				err := repositories.Repo.CreateTables(serverCfg.Storage, context.TODO())
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
