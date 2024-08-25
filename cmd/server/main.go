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
	"github.com/netzen86/collectmetrics/internal/handlers"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
)

const (
	addressServer    string = "localhost:8080"
	storeIntervalDef int    = 300
	metricFileName   string = "metrics.json"
)

func main() {
	var endpoint string
	var fileStoragePath string
	var storeInterval int
	var restore bool
	var err error
	flag.StringVar(&endpoint, "a", addressServer, "Used to set the address and port on which the server runs.")
	flag.StringVar(&fileStoragePath, "f", metricFileName, "Used to set file path to save metrics.")
	flag.BoolVar(&restore, "r", false, "Used to set restore metrics.")
	flag.IntVar(&storeInterval, "i", storeIntervalDef, "Used for set save metrics on disk.")

	flag.Parse()

	endpointTMP := os.Getenv("ADDRESS")
	if len(endpointTMP) != 0 {
		endpoint = endpointTMP
	}

	storeIntervalTmp := os.Getenv("STORE_INTERVAL")
	if len(storeIntervalTmp) != 0 {
		storeInterval, err = strconv.Atoi(storeIntervalTmp)
		if err != nil {
			fmt.Printf("%e\n", err)
			os.Exit(1)
		}
	}

	fileStoragePathTMP := os.Getenv("FILE_STORAGE_PATH")
	if len(fileStoragePathTMP) != 0 {
		fileStoragePath = fileStoragePathTMP
	}

	restoreTMP := os.Getenv("RESTORE")
	if len(restoreTMP) != 0 {
		restore, err = strconv.ParseBool(restoreTMP)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(flag.Args()) != 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	gw := chi.NewRouter()
	memSto, errm := memstorage.NewMemStorage()
	if errm != nil {
		panic(errm)
	}

	if restore {
		utils.LoadMetric(memSto, metricFileName)
	}

	ctx := context.Background()
	ms, err := memSto.GetMemStorage(ctx)
	if err != nil {
		panic(err)
	}

	gw.Route("/", func(gw chi.Router) {
		gw.Post("/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/", handlers.WithLogging(handlers.JSONUpdateMHandle(ms, fileStoragePath, storeInterval)))
		gw.Post("/value/", handlers.WithLogging(handlers.JSONRetrieveOneHandle(memSto)))
		gw.Post("/update/{mType}/{mName}", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/", handlers.WithLogging(handlers.BadRequest))
		gw.Post("/update/{mType}/{mName}/{mValue}", handlers.WithLogging(handlers.UpdateMHandle(memSto)))
		gw.Post("/*", handlers.WithLogging(handlers.NotFound))

		gw.Get("/value/{mType}/{mName}", handlers.WithLogging(handlers.RetrieveOneMHandle(memSto)))
		gw.Get("/", handlers.WithLogging(handlers.RetrieveMHandle(memSto)))
		gw.Get("/*", handlers.WithLogging(handlers.NotFound))

	},
	)

	if storeInterval != 0 {
		go utils.SaveMetrics(memSto, metricFileName, storeInterval)
	}

	errs := http.ListenAndServe(endpoint, gw)
	if errs != nil {
		panic(errs)
	}
}
