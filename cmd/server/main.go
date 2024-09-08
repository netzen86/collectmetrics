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
	"github.com/netzen86/collectmetrics/internal/db"
	"github.com/netzen86/collectmetrics/internal/handlers"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
)

const (
	addressServer    string = "localhost:8080"
	storeIntervalDef int    = 300
)

func main() {
	log.Println("!!! SERVER START !!!")
	var endpoint string
	var fileStoragePath string
	var dbconstring string
	var storeInterval int
	var tempfile *os.File
	var restore bool
	var err error

	storageSelecter := "MEMORY"
	saveMetricsDefaultPath := "servermetrics.json"

	workDir := utils.WorkingDir()
	// if !utils.ChkFileExist(workDir + saveMetricsDefaultPath) {
	// 	log.Fatal(err)
	// }
	saveMetricsDefaultPath = fmt.Sprintf("%s/%s", workDir, saveMetricsDefaultPath)

	flag.StringVar(&endpoint, "a", addressServer, "Used to set the address and port on which the server runs.")
	flag.StringVar(&fileStoragePath, "f", saveMetricsDefaultPath, "Used to set file path to save metrics.")
	flag.StringVar(&dbconstring, "d", "", "Used to set db connet string.")
	flag.BoolVar(&restore, "r", true, "Used to set restore metrics.")
	flag.IntVar(&storeInterval, "i", storeIntervalDef, "Used for set save metrics on disk.")

	flag.Parse()

	endpointTMP := os.Getenv("ADDRESS")
	if len(endpointTMP) != 0 {
		endpoint = os.Getenv("ADDRESS")
	}

	storeIntervalTmp := os.Getenv("STORE_INTERVAL")
	if len(storeIntervalTmp) != 0 {
		storeInterval, err = strconv.Atoi(os.Getenv("STORE_INTERVAL"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if fileStoragePath != saveMetricsDefaultPath && len(fileStoragePath) != 0 {
		saveMetricsDefaultPath = fileStoragePath
		storageSelecter = "FILE"
	}

	fileStoragePathTMP := os.Getenv("FILE_STORAGE_PATH")
	if len(fileStoragePathTMP) != 0 {
		fileStoragePath = os.Getenv("FILE_STORAGE_PATH")
		saveMetricsDefaultPath = fileStoragePath
		storageSelecter = "FILE"
	}

	if storageSelecter == "FILE" {
		log.Println("ENTER PRODUCER IN MAIN")
		_, err = os.MkdirTemp("", "tmp")
		if err != nil {
			log.Fatal(err)
		}
	}

	restoreTMP := os.Getenv("RESTORE")
	if len(restoreTMP) != 0 {
		restore, err = strconv.ParseBool(os.Getenv("RESTORE"))
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(dbconstring) != 0 {
		storageSelecter = "DATABASE"
	}

	dbaddressTMP := os.Getenv("DATABASE_DSN")
	if len(dbaddressTMP) != 0 {
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

	log.Println("!!! SERVER START !!!", endpoint, fileStoragePath, dbconstring, restore, storeInterval)

	gw := chi.NewRouter()

	memSto, err := memstorage.NewMemStorage()
	if err != nil {
		log.Fatal(err)
	}

	tempfile, err = os.OpenFile(fmt.Sprintf("%stmp", fileStoragePath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	if utils.ChkFileExist(fileStoragePath) {
		data, err := os.ReadFile(fileStoragePath) //read the contents of source file
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		err = os.WriteFile(fmt.Sprintf("%stmp", fileStoragePath), data, 0666) //write the content to destination file
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
	}

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
			memSto, tempfile, saveMetricsDefaultPath, dbconstring, storageSelecter, storeInterval)))
		gw.Post("/value/", handlers.WithLogging(handlers.JSONRetrieveOneHandle(
			memSto, tempfile.Name(), dbconstring, storageSelecter)))
		gw.Post("/update/{mType}/{mName}", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.WithLogging(handlers.UpdateMHandle(
			memSto, tempfile, saveMetricsDefaultPath, dbconstring, storageSelecter)))
		gw.Post("/*", handlers.WithLogging(handlers.NotFound))

		gw.Get("/ping", handlers.WithLogging(handlers.PingDB(dbconstring)))
		gw.Get("/value/{mType}/{mName}", handlers.WithLogging(handlers.RetrieveOneMHandle(
			memSto, fileStoragePath, dbconstring, storageSelecter)))
		gw.Get("/", handlers.WithLogging(handlers.RetrieveMHandle(memSto)))
		gw.Get("/*", handlers.WithLogging(handlers.NotFound))

	},
	)

	if storeInterval != 0 {
		go files.SaveMetrics(memSto, saveMetricsDefaultPath,
			tempfile.Name(), storageSelecter, storeInterval)
	}

	errs := http.ListenAndServe(endpoint, gw)
	if errs != nil {
		panic(errs)
	}
}
