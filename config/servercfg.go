// Package config
// Пакет для конфигурации приложения Сервер.
// Получает флаги и переменные окружения.
// Инициализирует некоторые функции.
package config

import (
	"bufio"
	"context"
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/db"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/security"
)

// константы используещиеся для работы Сервера.
const (
	addressServer    string = "localhost:8080"
	storeIntervalDef int    = 300
	// имена переменных окружения
	envAdd     string = "ADDRESS"
	envSI      string = "STORE_INTERVAL"
	envFSP     string = "FILE_STORAGE_PATH"
	envCFG     string = "CONFIG"
	envRes     string = "RESTORE"
	envKey     string = "KEY"
	envPRIVKEY string = "CRYPTO_KEY"
	envDB      string = "DATABASE_DSN"
)

type configSrvFile struct {
	Adderss   string `json:"address,omitempty"`        // аналог переменной окружения ADDRESS или флага -a
	Restore   bool   `json:"restore,omitempty"`        // аналог переменной окружения RESTORE или флага -r
	StorInter int    `json:"store_interval,omitempty"` // аналог переменной окружения STORE_INTERVAL или флага -i
	StoreFile string `json:"store_file,omitempty"`     // аналог переменной окружения STORE_FILE или -f
	Dsn       string `json:"database_dsn,omitempty"`   // аналог переменной окружения DATABASE_DSN или флага -d
	CryptoKey string `json:"crypto_key,omitempty"`     // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
}

// ServerCfg структура для конфигурации Сервера.
type ServerCfg struct {
	// указатель на memstorage
	Storage repositories.Repo `env:"" DefVal:""`
	// указатель на временный файл хранения метрик
	Tempfile *os.File `env:"" DefVal:""`
	// адрес и порт на котором запуститься сервер
	Endpoint string `env:"ADDRESS" DefVal:"localhost:8080"`
	// имя и путь к файлу для хранения метрик
	FileStoragePath string `env:"FILE_STORAGE_PATH" DefVal:""`
	// имя и путь к файлу для хранения метрик для значения по умолчанию
	FileStoragePathDef string `env:"" DefVal:"FileStoragePath"`
	// имя файла для получения конфигурации
	SrvFileCfg string `env:"" DefVal:""`
	// ключ для создания подписи данных
	SignKeyString string `env:"KEY" DefVal:""`
	// путь к файлу приватного ключа
	PrivKeyFileName string `env:"CRYPTO_KEY" DefVal:""`
	// приватный ключ для ассиметричного шифрования
	PrivKey *rsa.PrivateKey `env:"" DefVal:""`
	// если значенние флага true генерируем приватный и публичнные ключи
	KeyGenerate bool `env:"" DefVal:"false"`
	// строка для подключения к базе данных
	DBconstring string `env:"DATABASE_DSN" DefVal:""`
	// // ключ для выбора текущего хранилища (мемстораж, файл, база данных)
	// StorageSelecter string `env:"" DefVal:"MEMORY"`
	// интервал сохранения метрик в файл
	StoreInterval int `env:"STORE_INTERVAL" DefVal:"300s"`
	// ключ для определения восстановления метрик из файла
	Restore bool `env:"RESTORE" DefVal:"true"`
	// контекст для HTTP сервера
	ServerCtx context.Context `env:"" DefVal:""`
	// функция отмены для HTTP сервера
	ServerStopCtx context.CancelFunc `env:"" DefVal:""`
	// канал для os.Signal
	Sig chan os.Signal  `env:"" DefVal:""`
	Wg  *sync.WaitGroup `env:"" DefVal:""`
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
	flag.StringVar(&serverCfg.PrivKeyFileName, "crypto-key", "", "Load private key for decrypting.")
	flag.StringVar(&serverCfg.SrvFileCfg, "config", "", "Load configuration from file.")

	flag.BoolVar(&serverCfg.KeyGenerate, "g", false, "Used to generate private and public keys.")
	flag.BoolVar(&serverCfg.Restore, "r", true, "Used to set restore metrics.")
	flag.IntVar(&serverCfg.StoreInterval, "i", storeIntervalDef, "Used for set save metrics on disk.")

	flag.Parse()

	// если серверу преданы параменты, а не флаги
	// печатаем какие параметры можно использовать
	if len(flag.Args()) != 0 {
		flag.PrintDefaults()
		return fmt.Errorf("not args allowed")
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

	// получаем имя файла ключа для ассемитричного шифрования
	if len(os.Getenv(envPRIVKEY)) > 0 {
		serverCfg.PrivKeyFileName = os.Getenv(envPRIVKEY)
	}

	// получаем параметры подключения к базе данных
	if len(os.Getenv(envDB)) != 0 {
		serverCfg.DBconstring = os.Getenv(envDB)
	}
	return nil
}

// метод для получения параметров запуска сервера из файла формата json
func (serverCfg *ServerCfg) getSrvCfgFile() error {
	var srvCfg configSrvFile
	config, err := os.Open(serverCfg.SrvFileCfg)
	if err != nil {
		return fmt.Errorf("error when read server config file %w", err)
	}
	defer config.Close()

	fileinfo, _ := config.Stat()
	cfgBytes := make([]byte, fileinfo.Size())
	buffer := bufio.NewReader(config)
	_, err = buffer.Read(cfgBytes)
	if err != nil {
		return fmt.Errorf("error when read config file %w", err)
	}
	err = json.Unmarshal(cfgBytes, &srvCfg)
	if err != nil {
		return fmt.Errorf("error when unmarshal config %w", err)
	}
	switch {
	case serverCfg.Endpoint == addressServer:
		serverCfg.Endpoint = srvCfg.Adderss
		fallthrough
	case serverCfg.Restore:
		serverCfg.Restore = srvCfg.Restore
		fallthrough
	case serverCfg.StoreInterval == 300:
		serverCfg.StoreInterval = srvCfg.StorInter
		fallthrough
	case serverCfg.FileStoragePath == serverCfg.FileStoragePathDef:
		if len(srvCfg.StoreFile) != 0 {
			serverCfg.FileStoragePath = srvCfg.StoreFile
		}
		fallthrough
	case len(serverCfg.DBconstring) == 0:
		serverCfg.DBconstring = srvCfg.Dsn
		fallthrough
	case len(serverCfg.PrivKeyFileName) == 0:
		serverCfg.PrivKeyFileName = srvCfg.CryptoKey
	}
	return nil
}

// метод инициализации сервера
func (serverCfg *ServerCfg) initSrv(srvlog zap.SugaredLogger) error {
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
			return fmt.Errorf("error create temp file %w ", err)
		}
	}

	// создание приватного и публичного ключа
	if serverCfg.KeyGenerate {
		err = security.GenerateKeys()
		if err != nil {
			return fmt.Errorf("error generate rsa keys %w ", err)
		}
	}

	// считываем приваиный ключ
	if len(serverCfg.PrivKeyFileName) > 0 {
		serverCfg.PrivKey, err = security.ReadPrivedKey(security.PrivKeyFileName)
		if err != nil {
			return fmt.Errorf("error reading priv key file %w ", err)
		}
	}

	// создание контекста для graceful shutdown сервера
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	serverCfg.ServerCtx = serverCtx
	serverCfg.ServerStopCtx = serverStopCtx

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverCfg.Sig = sig

	serverCfg.Wg = &sync.WaitGroup{}

	// лог значений полученных из переменных окружения и флагов
	srvlog.Infoln("!!! SERVER CONFIGURED !!!",
		serverCfg.Endpoint, serverCfg.FileStoragePathDef,
		serverCfg.FileStoragePath, serverCfg.DBconstring,
		len(serverCfg.SignKeyString), serverCfg.Restore, serverCfg.StoreInterval)

	return nil
}

// GetServerCfg метод для получения конфигурации сервера
func (serverCfg *ServerCfg) GetServerCfg(srvlog zap.SugaredLogger) error {
	var err error

	err = serverCfg.parseSrvFlags()
	if err != nil {
		return fmt.Errorf("error parse flags: %w", err)
	}

	if len(serverCfg.SrvFileCfg) != 0 {
		err = serverCfg.getSrvCfgFile()
		if err != nil {
			return fmt.Errorf("error get config from file: %w", err)
		}
	}

	err = serverCfg.getSrvEnv()
	if err != nil {
		return fmt.Errorf("error get env var: %w", err)
	}

	err = serverCfg.initSrv(srvlog)
	if err != nil {
		return fmt.Errorf("error server init: %w", err)
	}
	return nil
}
