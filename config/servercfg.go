package config

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/netzen86/collectmetrics/internal/db"
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
	Endpoint           string
	FileStoragePath    string
	FileStoragePathDef string
	SignKeyString      string
	DBconstring        string
	StorageSelecter    string
	StoreInterval      int
	Restore            bool
	Tempfile           *os.File
	MemStorage         *memstorage.MemStorage // хранилище данных в памяти
}

func GetSrvEnv(srvcfg *ServerCfg) error {
	var err error

	// получаем данные для работы програмы из переменных окружения
	// переменные окружения имеют наивысший приоритет
	if len(os.Getenv(EnvAdd)) > 0 {
		srvcfg.Endpoint = os.Getenv(EnvAdd)
	}

	if len(os.Getenv(EnvSI)) != 0 {
		srvcfg.StoreInterval, err = strconv.Atoi(os.Getenv(EnvSI))
		if err != nil {
			return fmt.Errorf("error atoi poll interval %v ", err)
		}
	}

	if len(os.Getenv(EnvFSP)) != 0 {
		srvcfg.FileStoragePath = os.Getenv(EnvFSP)
		srvcfg.FileStoragePathDef = os.Getenv(EnvFSP)
		srvcfg.StorageSelecter = ssFile

	}

	// if storageSelecter == "FILE" {
	// 	log.Println("ENTER PRODUCER IN MAIN")
	// 	_, err = os.MkdirTemp("", "tmp")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	if len(os.Getenv(EnvRes)) != 0 {
		srvcfg.Restore, err = strconv.ParseBool(os.Getenv(EnvRes))
		if err != nil {
			return fmt.Errorf("error parse bool restore %v ", err)
		}
	}

	if len(os.Getenv(EnvKey)) != 0 {
		srvcfg.SignKeyString = os.Getenv(EnvKey)
	}

	if len(os.Getenv(EnvDB)) != 0 {
		srvcfg.DBconstring = os.Getenv(EnvDB)
		srvcfg.StorageSelecter = ssDataBase
	}
	return nil
}

func GetServerCfg() (ServerCfg, error) {
	var serverCfg ServerCfg
	var err error

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

	if serverCfg.FileStoragePath != serverCfg.FileStoragePathDef && len(serverCfg.FileStoragePath) != 0 {
		serverCfg.FileStoragePathDef = serverCfg.FileStoragePath
		serverCfg.StorageSelecter = ssFile
	}

	if len(serverCfg.DBconstring) != 0 {
		serverCfg.StorageSelecter = ssDataBase
	}

	err = GetSrvEnv(&serverCfg)
	if err != nil {
		return serverCfg, fmt.Errorf("error when get env var %v ", err)
	}

	serverCfg.MemStorage, err = memstorage.NewMemStorage()
	if err != nil {
		return serverCfg, fmt.Errorf("error when get mem storage %v ", err)
	}

	log.Println("!!! SERVER START !!!",
		serverCfg.Endpoint, serverCfg.FileStoragePathDef,
		serverCfg.FileStoragePath, serverCfg.DBconstring,
		len(serverCfg.SignKeyString), serverCfg.Restore, serverCfg.StoreInterval)

	serverCfg.Tempfile, err = os.OpenFile(fmt.Sprintf("%stmp", serverCfg.FileStoragePath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return serverCfg, fmt.Errorf("error create temp file %v ", err)
	}

	if utils.ChkFileExist(serverCfg.FileStoragePath) {
		data, err := os.ReadFile(serverCfg.FileStoragePath) //read the contents of source file
		if err != nil {
			return serverCfg, fmt.Errorf("error reading file: %v", err)
		}
		err = os.WriteFile(fmt.Sprintf("%stmp", serverCfg.FileStoragePath), data, 0666) //write the content to destination file
		if err != nil {
			return serverCfg, fmt.Errorf("error writing file: %v", err)
		}
	}

	if serverCfg.Restore {
		log.Println("ENTER IN RESTORE")
		files.LoadMetric(serverCfg.MemStorage, serverCfg.FileStoragePathDef)
	}

	if serverCfg.StorageSelecter == ssDataBase {
		retrybuilder := func() func() error {
			return func() error {
				err := db.CreateTables(context.TODO(), serverCfg.DBconstring)
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
	return serverCfg, nil
}
