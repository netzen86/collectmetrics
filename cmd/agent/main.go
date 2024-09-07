package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/netzen86/collectmetrics/internal/api"
	"github.com/netzen86/collectmetrics/internal/db"
	"github.com/netzen86/collectmetrics/internal/repositories"
	"github.com/netzen86/collectmetrics/internal/repositories/files"
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/utils"
	"github.com/spf13/pflag"
)

const (
	addressServer      string = "localhost:8080"
	templateAddressSrv string = "http://%s/update/"
	gag                string = "gauge"
	cnt                string = "counter"
	storefiledfl       string = "testagent.json"
	Alloc              string = "Alloc"
	BuckHashSys        string = "BuckHashSys"
	Frees              string = "Frees"
	GCCPUFraction      string = "GCCPUFraction"
	GCSys              string = "GCSys"
	HeapAlloc          string = "HeapAlloc"
	HeapIdle           string = "HeapIdle"
	HeapInuse          string = "HeapInuse"
	HeapObjects        string = "HeapObjects"
	HeapReleased       string = "HeapReleased"
	HeapSys            string = "HeapSys"
	LastGC             string = "LastGC"
	Lookups            string = "Lookups"
	MCacheInuse        string = "MCacheInuse"
	MCacheSys          string = "MCacheSys"
	MSpanInuse         string = "MSpanInuse"
	MSpanSys           string = "MSpanSys"
	Mallocs            string = "Mallocs"
	NextGC             string = "NextGC"
	NumForcedGC        string = "NumForcedGC"
	NumGC              string = "NumGC"
	OtherSys           string = "OtherSys"
	PauseTotalNs       string = "PauseTotalNs"
	StackInuse         string = "StackInuse"
	StackSys           string = "StackSys"
	Sys                string = "Sys"
	TotalAlloc         string = "TotalAlloc"
	PollCount          string = "PollCount"
	RandomValue        string = "RandomValue"
	pollInterval       int    = 2
	reportInterval     int    = 10
)

func CollectMetrics(storage repositories.Repo, tempfile *os.File, dbconstr, storageSelecter string, pollcnt int) {
	ctx := context.Background()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)

	if storageSelecter == "MEMORY" {
		storage.UpdateParam(ctx, false, gag, Alloc, float64(memStats.Alloc))
		storage.UpdateParam(ctx, false, gag, BuckHashSys, float64(memStats.BuckHashSys))
		storage.UpdateParam(ctx, false, gag, Frees, float64(memStats.Frees))
		storage.UpdateParam(ctx, false, gag, GCCPUFraction, float64(memStats.GCCPUFraction))
		storage.UpdateParam(ctx, false, gag, GCSys, float64(memStats.GCSys))
		storage.UpdateParam(ctx, false, gag, HeapAlloc, float64(memStats.HeapAlloc))
		storage.UpdateParam(ctx, false, gag, HeapIdle, float64(memStats.HeapIdle))
		storage.UpdateParam(ctx, false, gag, HeapInuse, float64(memStats.HeapInuse))
		storage.UpdateParam(ctx, false, gag, HeapObjects, float64(memStats.HeapObjects))
		storage.UpdateParam(ctx, false, gag, HeapReleased, float64(memStats.HeapReleased))
		storage.UpdateParam(ctx, false, gag, HeapSys, float64(memStats.HeapSys))
		storage.UpdateParam(ctx, false, gag, LastGC, float64(memStats.LastGC))
		storage.UpdateParam(ctx, false, gag, Lookups, float64(memStats.Lookups))
		storage.UpdateParam(ctx, false, gag, MCacheInuse, float64(memStats.MCacheInuse))
		storage.UpdateParam(ctx, false, gag, MCacheSys, float64(memStats.MCacheSys))
		storage.UpdateParam(ctx, false, gag, MSpanInuse, float64(memStats.MSpanInuse))
		storage.UpdateParam(ctx, false, gag, Mallocs, float64(memStats.Mallocs))
		storage.UpdateParam(ctx, false, gag, MSpanSys, float64(memStats.MSpanSys))
		storage.UpdateParam(ctx, false, gag, NextGC, float64(memStats.NextGC))
		storage.UpdateParam(ctx, false, gag, NumForcedGC, float64(memStats.NumForcedGC))
		storage.UpdateParam(ctx, false, gag, NumGC, float64(memStats.NumGC))
		storage.UpdateParam(ctx, false, gag, OtherSys, float64(memStats.OtherSys))
		storage.UpdateParam(ctx, false, gag, PauseTotalNs, float64(memStats.PauseTotalNs))
		storage.UpdateParam(ctx, false, gag, StackInuse, float64(memStats.StackInuse))
		storage.UpdateParam(ctx, false, gag, StackSys, float64(memStats.StackSys))
		storage.UpdateParam(ctx, false, gag, Sys, float64(memStats.Sys))
		storage.UpdateParam(ctx, false, gag, TotalAlloc, float64(memStats.TotalAlloc))
		storage.UpdateParam(ctx, false, gag, RandomValue, rand.Float64())
		storage.UpdateParam(ctx, false, cnt, PollCount, int64(1))
	}

	if storageSelecter == "DATABASE" {
		db.UpdateParamDB(ctx, dbconstr, gag, Alloc, float64(memStats.Alloc))
		db.UpdateParamDB(ctx, dbconstr, gag, BuckHashSys, float64(memStats.BuckHashSys))
		db.UpdateParamDB(ctx, dbconstr, gag, Frees, float64(memStats.Frees))
		db.UpdateParamDB(ctx, dbconstr, gag, GCCPUFraction, float64(memStats.GCCPUFraction))
		db.UpdateParamDB(ctx, dbconstr, gag, GCSys, float64(memStats.GCSys))
		db.UpdateParamDB(ctx, dbconstr, gag, HeapAlloc, float64(memStats.HeapAlloc))
		db.UpdateParamDB(ctx, dbconstr, gag, HeapIdle, float64(memStats.HeapIdle))
		db.UpdateParamDB(ctx, dbconstr, gag, HeapInuse, float64(memStats.HeapInuse))
		db.UpdateParamDB(ctx, dbconstr, gag, HeapObjects, float64(memStats.HeapObjects))
		db.UpdateParamDB(ctx, dbconstr, gag, HeapReleased, float64(memStats.HeapReleased))
		db.UpdateParamDB(ctx, dbconstr, gag, HeapSys, float64(memStats.HeapSys))
		db.UpdateParamDB(ctx, dbconstr, gag, LastGC, float64(memStats.LastGC))
		db.UpdateParamDB(ctx, dbconstr, gag, Lookups, float64(memStats.Lookups))
		db.UpdateParamDB(ctx, dbconstr, gag, MCacheInuse, float64(memStats.MCacheInuse))
		db.UpdateParamDB(ctx, dbconstr, gag, MCacheSys, float64(memStats.MCacheSys))
		db.UpdateParamDB(ctx, dbconstr, gag, MSpanInuse, float64(memStats.MSpanInuse))
		db.UpdateParamDB(ctx, dbconstr, gag, Mallocs, float64(memStats.Mallocs))
		db.UpdateParamDB(ctx, dbconstr, gag, MSpanSys, float64(memStats.MSpanSys))
		db.UpdateParamDB(ctx, dbconstr, gag, NextGC, float64(memStats.NextGC))
		db.UpdateParamDB(ctx, dbconstr, gag, NumForcedGC, float64(memStats.NumForcedGC))
		db.UpdateParamDB(ctx, dbconstr, gag, NumGC, float64(memStats.NumGC))
		db.UpdateParamDB(ctx, dbconstr, gag, OtherSys, float64(memStats.OtherSys))
		db.UpdateParamDB(ctx, dbconstr, gag, PauseTotalNs, float64(memStats.PauseTotalNs))
		db.UpdateParamDB(ctx, dbconstr, gag, StackInuse, float64(memStats.StackInuse))
		db.UpdateParamDB(ctx, dbconstr, gag, StackSys, float64(memStats.StackSys))
		db.UpdateParamDB(ctx, dbconstr, gag, Sys, float64(memStats.Sys))
		db.UpdateParamDB(ctx, dbconstr, gag, TotalAlloc, float64(memStats.TotalAlloc))
		db.UpdateParamDB(ctx, dbconstr, gag, RandomValue, rand.Float64())
		db.UpdateParamDB(ctx, dbconstr, cnt, PollCount, int64(1))
	}
	if storageSelecter == "FILE" {
		if err := os.Truncate(tempfile.Name(), 0); err != nil {
			log.Printf("Failed to truncate: %v", err)
		}
		alloc := float64(memStats.Alloc)
		buckHashSys := float64(memStats.BuckHashSys)
		frees := float64(memStats.Frees)
		gccpuFraction := float64(memStats.GCCPUFraction)
		gcSys := float64(memStats.GCSys)
		heapAlloc := float64(memStats.HeapAlloc)
		heapIdle := float64(memStats.HeapIdle)
		heapInuse := float64(memStats.HeapInuse)
		heapObjects := float64(memStats.HeapObjects)
		heapReleased := float64(memStats.HeapReleased)
		heapSys := float64(memStats.HeapSys)
		lastGC := float64(memStats.LastGC)
		lookups := float64(memStats.Lookups)
		mCacheInuse := float64(memStats.MCacheInuse)
		mCacheSys := float64(memStats.MCacheSys)
		mSpanInuse := float64(memStats.MSpanInuse)
		mallocs := float64(memStats.Mallocs)
		mSpanSys := float64(memStats.MSpanSys)
		nextGC := float64(memStats.NextGC)
		numForcedGC := float64(memStats.NumForcedGC)
		numGC := float64(memStats.NumGC)
		otherSys := float64(memStats.OtherSys)
		pauseTotalNs := float64(memStats.PauseTotalNs)
		stackInuse := float64(memStats.StackInuse)
		stackSys := float64(memStats.StackSys)
		sys := float64(memStats.Sys)
		totalAlloc := float64(memStats.TotalAlloc)
		randomValue := rand.Float64()
		pollCount := int64(pollcnt)

		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: Alloc, Value: &alloc})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: BuckHashSys, Value: &buckHashSys})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: Frees, Value: &frees})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: GCCPUFraction, Value: &gccpuFraction})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: GCSys, Value: &gcSys})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: HeapAlloc, Value: &heapAlloc})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: HeapIdle, Value: &heapIdle})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: HeapInuse, Value: &heapInuse})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: HeapObjects, Value: &heapObjects})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: HeapReleased, Value: &heapReleased})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: HeapSys, Value: &heapSys})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: LastGC, Value: &lastGC})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: Lookups, Value: &lookups})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: MCacheInuse, Value: &mCacheInuse})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: MCacheSys, Value: &mCacheSys})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: MSpanInuse, Value: &mSpanInuse})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: Mallocs, Value: &mallocs})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: MSpanSys, Value: &mSpanSys})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: NextGC, Value: &nextGC})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: NumForcedGC, Value: &numForcedGC})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: NumGC, Value: &numGC})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: OtherSys, Value: &otherSys})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: PauseTotalNs, Value: &pauseTotalNs})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: StackInuse, Value: &stackInuse})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: StackSys, Value: &stackSys})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: Sys, Value: &sys})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: TotalAlloc, Value: &totalAlloc})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: gag, ID: RandomValue, Value: &randomValue})
		files.FileStorage(ctx, tempfile, api.Metrics{MType: cnt, ID: PollCount, Delta: &pollCount})
	}
}

func SendMetrics(url, metricData string) error {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", url, metricData), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", api.Th)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	return nil
}

func GetAccEnc(url, contEnc string) (string, error) {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	request.Header.Add("Content-Encoding", contEnc)
	request.Header.Add("Content-Type", api.Js)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	defer response.Body.Close()

	log.Printf("%s %s\n", "GetAccEnc", response.Status)
	encoding := response.Header.Get("Accept-Encoding")
	return encoding, nil
}

func JSONdecode(resp *http.Response) {
	var buf bytes.Buffer
	var metrics api.Metrics
	if resp == nil {
		log.Print("error nil response")
		return
	}
	defer resp.Body.Close()
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		log.Print("reading body error ", err)
		return
	}
	// если данные запакованные
	err = utils.SelectDeCoHTTP(&buf, resp)
	if err != nil {
		log.Print("unpack data error", err)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		log.Print("parse json error", err)
		return
	}
	if metrics.MType == "counter" {
		log.Printf("%s %v\n", metrics.ID, *metrics.Delta)
	}
	if metrics.MType == "gauge" {
		log.Printf("%s %v\n", metrics.ID, *metrics.Value)
	}
}

func JSONSendMetrics(url, ce string, metricsData api.Metrics) (*http.Response, error) {
	// получаем от сервера ответ о поддерживаемыж методах сжатия
	encoding, err := GetAccEnc(url, ce)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	// сериализуем данные в JSON
	data, err := json.Marshal(metricsData)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	metricsData.Clean()
	// если сервер поддерживает сжатие сжимаем данные
	if encoding == "gzip" {
		data, err = utils.GzipCompress(data)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	if encoding == "gzip" {
		request.Header.Add("Content-Encoding", api.Gz)
	}

	request.Header.Add("Content-Type", api.Js)
	request.Header.Add("Accept-Encoding", api.Gz)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	if response.StatusCode != 200 {
		return nil, errors.New(response.Status)
	}
	// defer response.Body.Close()
	return response, nil
}

func CommonSendGag(nojson bool, endpoint, contentEnc, key string, value float64) {
	if nojson {
		err := SendMetrics(fmt.Sprintf(templateAddressSrv, endpoint), fmt.Sprintf("gauge/%s/%v", key, value))
		if err != nil {
			log.Println(err)
		}
	} else if !nojson {
		resp, err := JSONSendMetrics(
			fmt.Sprintf(templateAddressSrv, endpoint),
			contentEnc,
			api.Metrics{MType: "gauge", ID: key, Value: &value})
		if err != nil {
			log.Println("Gauge error ", err)
		}
		JSONdecode(resp)
	}
}

func CommonSendCnt(nojson bool, endpoint, contentEnc, key string, value int64) {
	if nojson {
		err := SendMetrics(fmt.Sprintf(templateAddressSrv, endpoint), fmt.Sprintf("counter/%s/%v", key, value))
		if err != nil {
			log.Println(err)
		}
	} else if !nojson {
		resp, err := JSONSendMetrics(
			fmt.Sprintf(templateAddressSrv, endpoint),
			contentEnc,
			api.Metrics{MType: "counter", ID: key, Delta: &value})
		if err != nil {
			log.Println(err)
		}
		JSONdecode(resp)
	}
}

func iterMemStorage(storage *memstorage.MemStorage, nojson bool, endpoint, contentEnc string) {
	for k, v := range storage.Gauge {
		CommonSendGag(nojson, endpoint, contentEnc, k, v)
	}
	for k, v := range storage.Counter {
		CommonSendCnt(nojson, endpoint, contentEnc, k, v)
	}
}

func iterDB(nojson bool, dbconstr, endpoint, contentEnc string) {
	var metric api.Metrics

	db, err := db.ConDB(dbconstr)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	smtpGag := `SELECT name, value FROM gauge`
	rows, err := db.QueryContext(context.TODO(), smtpGag)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&metric.ID, &metric.Value)
		if err != nil {
			log.Println(err)
		}
		CommonSendGag(nojson, endpoint, contentEnc, metric.ID, *metric.Value)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	smtpCnt := `SELECT name, delta FROM counter`
	rows, err = db.QueryContext(context.TODO(), smtpCnt)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&metric.ID, &metric.Delta)
		if err != nil {
			log.Println(err)
		}
		CommonSendCnt(nojson, endpoint, contentEnc, metric.ID, *metric.Delta)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
}

func iterFile(nojson bool, fileStoragePath, endpoint, contentEnc string) {
	metric := api.Metrics{}
	consumer, err := files.NewConsumer(fileStoragePath)
	if err != nil {
		log.Println(err, " can't create consumer in if")
	}

	scanner := consumer.Scanner
	for scanner.Scan() {
		// преобразуем данные из JSON-представления в структуру
		err := json.Unmarshal(scanner.Bytes(), &metric)
		if err != nil {
			log.Printf("can't unmarshal string %v", err)
		}
		if metric.MType == "gauge" {
			CommonSendGag(nojson, endpoint, contentEnc, metric.ID, *metric.Value)
		}
		if metric.MType == "counter" {
			CommonSendCnt(nojson, endpoint, contentEnc, metric.ID, *metric.Delta)
		}
	}
}

func main() {
	var endpoint string
	var contentEnc string
	var dbconstring string
	var fileStoragePath string
	var nojson bool
	var pInterv int
	var rInterv int
	var err error
	storageSelecter := "MEMORY"
	ctx := context.TODO()
	// устанвливаем для отображения даты и времени в логах
	log.SetFlags(log.Ldate | log.Ltime)

	// опредаляем флаги
	pflag.StringVarP(&endpoint, "endpoint", "a", addressServer, "Used to set the address and port to connect server.")
	pflag.StringVarP(&contentEnc, "contentenc", "c", api.Gz, "Used to set content encoding to connect server.")
	pflag.StringVarP(&fileStoragePath, "filepath", "f", storefiledfl, "Used to set file path to save metrics.")
	pflag.StringVarP(&dbconstring, "dbconstring", "d", "", "Used to set file path to save metrics.")
	pflag.IntVarP(&pInterv, "pollinterval", "p", pollInterval, "User for set poll interval in seconds.")
	pflag.IntVarP(&rInterv, "reportinterval", "r", reportInterval, "User for set report interval (send to srv) in seconds.")
	pflag.BoolVarP(&nojson, "nojson", "n", false, "Use for enable url request")
	pflag.Parse()

	// если переданы аргументы не флаги печатаем подсказку
	if len(pflag.Args()) != 0 {
		pflag.PrintDefaults()
		os.Exit(1)
	}
	// получаем данные для работы програмы из переменных окружения
	// переменные окружения имеют наивысший приоритет
	endpointTMP := os.Getenv("ADDRESS")
	if len(endpointTMP) != 0 {
		endpoint = endpointTMP
	}

	pIntervTmp := os.Getenv("POLL_INTERVAL")
	if len(pIntervTmp) != 0 {
		pInterv, err = strconv.Atoi(pIntervTmp)
		if err != nil {
			fmt.Printf("%e\n", err)
			os.Exit(1)
		}
	}

	rIntervTmp := os.Getenv("REPORT_INTERVAL")
	if len(rIntervTmp) != 0 {
		rInterv, err = strconv.Atoi(rIntervTmp)
		if err != nil {
			fmt.Printf("%e\n", err)
			os.Exit(1)
		}
	}

	if len(fileStoragePath) != 0 {
		storageSelecter = "FILE"
	}

	fileStoragePathTMP := os.Getenv("FILE_STORAGE_PATH")
	if len(fileStoragePathTMP) != 0 {
		fileStoragePath = fileStoragePathTMP
		storageSelecter = "FILE"
	}

	if len(dbconstring) != 0 {
		storageSelecter = "DATABASE"
	}

	dbaddressTMP := os.Getenv("DATABASE_DSN")
	if len(dbaddressTMP) != 0 {
		dbconstring = dbaddressTMP
		storageSelecter = "DATABASE"
	}

	if storageSelecter == "DATABASE" {
		err = db.CreateTables(ctx, dbconstring)
		if err != nil {
			log.Println(err)
		}
	}

	storage, err := memstorage.NewMemStorage()
	if err != nil {
		panic("couldn't alloc mem")
	}

	storefile, err := os.OpenFile(fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}

	pollTik := time.NewTicker(time.Duration(pInterv) * time.Second)
	reportTik := time.NewTicker(time.Duration(rInterv) * time.Second)
	counter := 0
	for {
		select {
		case <-pollTik.C:
			counter += 1
			CollectMetrics(storage, storefile, fileStoragePath, storageSelecter, counter)
		case <-reportTik.C:
			if storageSelecter == "MEMORY" {
				iterMemStorage(storage, nojson, endpoint, contentEnc)
			}
			if storageSelecter == "DATABASE" {
				iterDB(nojson, dbconstring, endpoint, contentEnc)
			}
			if storageSelecter == "FILE" {
				iterFile(nojson, fileStoragePath, endpoint, contentEnc)
			}
		}
	}
}
