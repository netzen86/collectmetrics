package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/db"
	"github.com/netzen86/collectmetrics/internal/handlers"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

const (
	addressServer    string = "localhost:8080"
	storeIntervalDef int    = 300
)

func main() {
	var endpoint string
	var fileStoragePath string
	var dbconstring string
	var storeInterval int
	var producer *files.Producer
	var pcMetric api.Metrics
	var restore bool
	var err error

	storageSelecter := "MEMORY"
	saveMetricsDefaultPath := "servermetrics.json"

	flag.StringVar(&endpoint, "a", addressServer, "Used to set the address and port on which the server runs.")
	flag.StringVar(&fileStoragePath, "f", saveMetricsDefaultPath, "Used to set file path to save metrics.")
	flag.StringVar(&dbconstring, "d", "", "Used to set db connet string.")
	flag.BoolVar(&restore, "r", true, "Used to set restore metrics.")
	flag.IntVar(&storeInterval, "i", storeIntervalDef, "Used for set save metrics on disk.")

	flag.Parse()

	// endpointTMP := os.Getenv("ADDRESS")
	if len(os.Getenv("ADDRESS")) != 0 {
		endpoint = os.Getenv("ADDRESS")
	}

	// storeIntervalTmp := os.Getenv("STORE_INTERVAL")
	if len(os.Getenv("STORE_INTERVAL")) != 0 {
		storeInterval, err = strconv.Atoi(os.Getenv("STORE_INTERVAL"))
		if err != nil {
			fmt.Printf("%e\n", err)
			os.Exit(1)
		}
	}

	if fileStoragePath != saveMetricsDefaultPath && len(fileStoragePath) != 0 {
		saveMetricsDefaultPath = fileStoragePath
		storageSelecter = "FILE"
	}

	// fileStoragePathTMP := os.Getenv("FILE_STORAGE_PATH")
	if len(os.Getenv("FILE_STORAGE_PATH")) != 0 {
		fileStoragePath = os.Getenv("FILE_STORAGE_PATH")
		saveMetricsDefaultPath = fileStoragePath
		storageSelecter = "FILE"
	}

	if storageSelecter == "FILE" {
		log.Println("ENTER PRODUCER IN MAIN")
		producer, err = files.NewProducer(fileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	// restoreTMP := os.Getenv("RESTORE")
	if len(os.Getenv("RESTORE")) != 0 {
		restore, err = strconv.ParseBool(os.Getenv("RESTORE"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(dbconstring) != 0 {
		storageSelecter = "DATABASE"
	}

	// dbaddressTMP := os.Getenv("DATABASE_DSN")
	if len(os.Getenv("DATABASE_DSN")) != 0 {
		dbconstring = os.Getenv("DATABASE_DSN")
		storageSelecter = "DATABASE"
		err = db.CreateTables(context.TODO(), dbconstring)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(flag.Args()) != 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	gw := chi.NewRouter()

	memSto, err := memstorage.NewMemStorage()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("!!!SERVER", storageSelecter, restore)

	if restore {
		log.Println("ENTER IN RESTORE")
		files.LoadMetric(memSto, saveMetricsDefaultPath)
	}

	if storageSelecter == "DATABASE" {
		err = db.CreateTables(context.TODO(), dbconstring)
		if err != nil {
			log.Fatal(err)
		}
	}

	gw.Route("/", func(gw chi.Router) {
		gw.Post("/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/", handlers.WithLogging(handlers.JSONUpdateMHandle(
			memSto, &pcMetric, producer, saveMetricsDefaultPath, dbconstring, storageSelecter, storeInterval)))
		gw.Post("/value/", handlers.WithLogging(handlers.JSONRetrieveOneHandle(
			memSto, fileStoragePath, dbconstring, storageSelecter)))
		gw.Post("/update/{mType}/{mName}", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.WithLogging(handlers.UpdateMHandle(
			memSto, &pcMetric, producer, dbconstring, storageSelecter)))
		gw.Post("/*", handlers.WithLogging(handlers.NotFound))

		gw.Get("/ping", handlers.WithLogging(handlers.PingDB(dbconstring)))
		gw.Get("/value/{mType}/{mName}", handlers.WithLogging(handlers.RetrieveOneMHandle(
			memSto, fileStoragePath, dbconstring, storageSelecter)))
		gw.Get("/", handlers.WithLogging(handlers.RetrieveMHandle(memSto)))
		gw.Get("/*", handlers.WithLogging(handlers.NotFound))

	},
	)

	if storeInterval != 0 {
		go files.SaveMetrics(memSto, saveMetricsDefaultPath, storeInterval)
	}

	errs := http.ListenAndServe(endpoint, gw)
	if errs != nil {
		panic(errs)
	}
}
