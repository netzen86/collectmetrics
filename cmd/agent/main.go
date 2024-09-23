package main

import (
	"bytes"
	"context"
	"encoding/hex"
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
	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
	"github.com/netzen86/collectmetrics/internal/security"
	"github.com/netzen86/collectmetrics/internal/utils"
	"github.com/spf13/pflag"
)

const (
	addressServer      string = "localhost:8080"
	templateAddressSrv string = "http://%s/update/"
	updatesAddress     string = "http://%s/updates/"
	gag                string = "gauge"
	cnt                string = "counter"
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

func CollectMetrics(storage repositories.Repo, tempfilename, dbconstr, storageSelecter string, pollcnt int) error {
	ctx := context.Background()
	var memStats runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&memStats)

	if storageSelecter == "MEMORY" {
		err := storage.UpdateParam(ctx, false, gag, Alloc, float64(memStats.Alloc))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, BuckHashSys, float64(memStats.BuckHashSys))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, Frees, float64(memStats.Frees))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, GCCPUFraction, float64(memStats.GCCPUFraction))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, GCSys, float64(memStats.GCSys))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, HeapAlloc, float64(memStats.HeapAlloc))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, HeapIdle, float64(memStats.HeapIdle))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, HeapInuse, float64(memStats.HeapInuse))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, HeapObjects, float64(memStats.HeapObjects))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, HeapReleased, float64(memStats.HeapReleased))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, HeapSys, float64(memStats.HeapSys))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, LastGC, float64(memStats.LastGC))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, Lookups, float64(memStats.Lookups))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, MCacheInuse, float64(memStats.MCacheInuse))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, MCacheSys, float64(memStats.MCacheSys))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, MSpanInuse, float64(memStats.MSpanInuse))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, Mallocs, float64(memStats.Mallocs))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, MSpanSys, float64(memStats.MSpanSys))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, NextGC, float64(memStats.NextGC))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, NumForcedGC, float64(memStats.NumForcedGC))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, NumGC, float64(memStats.NumGC))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, OtherSys, float64(memStats.OtherSys))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, PauseTotalNs, float64(memStats.PauseTotalNs))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, StackInuse, float64(memStats.StackInuse))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, StackSys, float64(memStats.StackSys))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, Sys, float64(memStats.Sys))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, TotalAlloc, float64(memStats.TotalAlloc))
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, gag, RandomValue, rand.Float64())
		if err != nil {
			return err
		}
		err = storage.UpdateParam(ctx, false, cnt, PollCount, int64(1))
		if err != nil {
			return err
		}
	}

	if storageSelecter == "DATABASE" {
		err := db.UpdateParamDB(ctx, dbconstr, gag, Alloc, float64(memStats.Alloc))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, BuckHashSys, float64(memStats.BuckHashSys))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, Frees, float64(memStats.Frees))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, GCCPUFraction, float64(memStats.GCCPUFraction))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, GCSys, float64(memStats.GCSys))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, HeapAlloc, float64(memStats.HeapAlloc))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, HeapIdle, float64(memStats.HeapIdle))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, HeapInuse, float64(memStats.HeapInuse))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, HeapObjects, float64(memStats.HeapObjects))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, HeapReleased, float64(memStats.HeapReleased))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, HeapSys, float64(memStats.HeapSys))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, LastGC, float64(memStats.LastGC))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, Lookups, float64(memStats.Lookups))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, MCacheInuse, float64(memStats.MCacheInuse))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, MCacheSys, float64(memStats.MCacheSys))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, MSpanInuse, float64(memStats.MSpanInuse))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, Mallocs, float64(memStats.Mallocs))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, MSpanSys, float64(memStats.MSpanSys))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, NextGC, float64(memStats.NextGC))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, NumForcedGC, float64(memStats.NumForcedGC))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, NumGC, float64(memStats.NumGC))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, OtherSys, float64(memStats.OtherSys))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, PauseTotalNs, float64(memStats.PauseTotalNs))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, StackInuse, float64(memStats.StackInuse))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, StackSys, float64(memStats.StackSys))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, Sys, float64(memStats.Sys))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, TotalAlloc, float64(memStats.TotalAlloc))
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, gag, RandomValue, rand.Float64())
		if err != nil {
			return err
		}
		err = db.UpdateParamDB(ctx, dbconstr, cnt, PollCount, int64(1))
		if err != nil {
			return err
		}
	}
	// if storageSelecter == "FILE" {
	// 	if err := os.Truncate(tempfilename, 0); err != nil {
	// 		log.Printf("Failed to truncate: %v", err)
	// 	}
	// 	alloc := float64(memStats.Alloc)
	// 	buckHashSys := float64(memStats.BuckHashSys)
	// 	frees := float64(memStats.Frees)
	// 	gccpuFraction := float64(memStats.GCCPUFraction)
	// 	gcSys := float64(memStats.GCSys)
	// 	heapAlloc := float64(memStats.HeapAlloc)
	// 	heapIdle := float64(memStats.HeapIdle)
	// 	heapInuse := float64(memStats.HeapInuse)
	// 	heapObjects := float64(memStats.HeapObjects)
	// 	heapReleased := float64(memStats.HeapReleased)
	// 	heapSys := float64(memStats.HeapSys)
	// 	lastGC := float64(memStats.LastGC)
	// 	lookups := float64(memStats.Lookups)
	// 	mCacheInuse := float64(memStats.MCacheInuse)
	// 	mCacheSys := float64(memStats.MCacheSys)
	// 	mSpanInuse := float64(memStats.MSpanInuse)
	// 	mallocs := float64(memStats.Mallocs)
	// 	mSpanSys := float64(memStats.MSpanSys)
	// 	nextGC := float64(memStats.NextGC)
	// 	numForcedGC := float64(memStats.NumForcedGC)
	// 	numGC := float64(memStats.NumGC)
	// 	otherSys := float64(memStats.OtherSys)
	// 	pauseTotalNs := float64(memStats.PauseTotalNs)
	// 	stackInuse := float64(memStats.StackInuse)
	// 	stackSys := float64(memStats.StackSys)
	// 	sys := float64(memStats.Sys)
	// 	totalAlloc := float64(memStats.TotalAlloc)
	// 	randomValue := rand.Float64()
	// 	pollCount := int64(pollcnt)

	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: Alloc, Value: &alloc})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: BuckHashSys, Value: &buckHashSys})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: Frees, Value: &frees})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: GCCPUFraction, Value: &gccpuFraction})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: GCSys, Value: &gcSys})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: HeapAlloc, Value: &heapAlloc})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: HeapIdle, Value: &heapIdle})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: HeapInuse, Value: &heapInuse})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: HeapObjects, Value: &heapObjects})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: HeapReleased, Value: &heapReleased})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: HeapSys, Value: &heapSys})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: LastGC, Value: &lastGC})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: Lookups, Value: &lookups})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: MCacheInuse, Value: &mCacheInuse})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: MCacheSys, Value: &mCacheSys})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: MSpanInuse, Value: &mSpanInuse})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: Mallocs, Value: &mallocs})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: MSpanSys, Value: &mSpanSys})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: NextGC, Value: &nextGC})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: NumForcedGC, Value: &numForcedGC})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: NumGC, Value: &numGC})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: OtherSys, Value: &otherSys})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: PauseTotalNs, Value: &pauseTotalNs})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: StackInuse, Value: &stackInuse})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: StackSys, Value: &stackSys})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: Sys, Value: &sys})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: TotalAlloc, Value: &totalAlloc})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: gag, ID: RandomValue, Value: &randomValue})
	// 	files.FileStorage(ctx, tempfilename, api.Metrics{MType: cnt, ID: PollCount, Delta: &pollCount})
	// }
	return nil
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

func ChkUpdates(endpoint string) bool {
	batchSend := false

	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf(updatesAddress, endpoint), nil)
	if err != nil {
		log.Printf("chk updates create req err %v", err)
		return batchSend
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("chk updates do req err %v", err)
		return batchSend
	}
	defer response.Body.Close()

	log.Printf("%s %s\n", "ChkUpdates", response.Status)

	if response.StatusCode == 200 {
		batchSend = true
	}
	return batchSend
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

func JSONdecode(resp *http.Response, batchSend bool) {
	var buf bytes.Buffer
	var metric api.Metrics
	var metrics []api.Metrics
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
	switch {
	case batchSend:
		if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
			log.Print("parse json error ", err)
			return
		}
		for _, m := range metrics {
			if m.MType == "counter" {
				log.Printf("%s %v\n", m.ID, *m.Delta)
			}
			if m.MType == "gauge" {
				log.Printf("%s %v\n", m.ID, *m.Value)
			}
		}
	default:
		if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
			log.Print("parse json error ", err)
			return
		}
		if metric.MType == "counter" {
			log.Printf("%s %v\n", metric.ID, *metric.Delta)
		}
		if metric.MType == "gauge" {
			log.Printf("%s %v\n", metric.ID, *metric.Value)
		}
	}
}

func JSONSendMetrics(url, ce, key string, metricsData api.Metrics, metrics []api.Metrics) (*http.Response, error) {
	var data, sing []byte

	// получаем от сервера ответ о поддерживаемыж методах сжатия
	encoding, err := GetAccEnc(url, ce)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	// сериализуем данные в JSON
	switch {
	case len(metrics) > 0:
		data, err = json.Marshal(metrics)
		if err != nil {
			log.Printf("serilazing error: %v\n", err)
			return nil, fmt.Errorf("serilazing error: %v", err)
		}
	default:
		data, err = json.Marshal(metricsData)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		metricsData.Clean()
	}
	if len(key) != 0 {
		sing = security.SingSendData(data, []byte(key))
	}

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
	if len(key) != 0 {
		request.Header.Add("HashSHA256", hex.EncodeToString(sing))
	}

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

func CommonSendGag(nojson bool, endpoint, contentEnc, key, singKey string, value float64) error {
	var metrics []api.Metrics
	if nojson {
		err := SendMetrics(fmt.Sprintf(templateAddressSrv, endpoint), fmt.Sprintf("gauge/%s/%v", key, value))
		if err != nil {
			return err
		}
	} else if !nojson {
		resp, err := JSONSendMetrics(
			fmt.Sprintf(templateAddressSrv, endpoint),
			contentEnc, singKey,
			api.Metrics{MType: "gauge", ID: key, Value: &value},
			metrics)
		if err != nil {
			return fmt.Errorf("gauge error %v", err)
		}
		JSONdecode(resp, false)
	}
	return nil
}

func CommonSendCnt(nojson bool, endpoint, contentEnc, key, singKey string, value int64) error {
	var metrics []api.Metrics
	if nojson {
		err := SendMetrics(fmt.Sprintf(templateAddressSrv, endpoint), fmt.Sprintf("counter/%s/%v", key, value))
		if err != nil {
			return err
		}
	} else if !nojson {
		resp, err := JSONSendMetrics(
			fmt.Sprintf(templateAddressSrv, endpoint),
			contentEnc, singKey,
			api.Metrics{MType: "counter", ID: key, Delta: &value},
			metrics)
		if err != nil {
			return fmt.Errorf("count error %v", err)

		}
		JSONdecode(resp, false)
	}
	return nil
}

func iterMemStorage(storage *memstorage.MemStorage, nojson, batchSend bool, endpoint, contentEnc, singKey string) {
	var metrics []api.Metrics
	for k, v := range storage.Gauge {
		switch {
		case batchSend:
			metrics = append(metrics, api.Metrics{MType: "gauge", ID: k, Value: &v})
		default:
			retrybuilder := func() func() error {
				return func() error {
					err := CommonSendGag(nojson, endpoint, contentEnc, singKey, k, v)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				log.Fatal("send gauge mertic fail ", err)
			}
		}
	}
	for k, v := range storage.Counter {
		switch {
		case batchSend:
			metrics = append(metrics, api.Metrics{MType: "counter", ID: k, Delta: &v})
		default:
			retrybuilder := func() func() error {
				return func() error {
					err := CommonSendCnt(nojson, endpoint, contentEnc, singKey, k, v)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				log.Fatal("send counter mertic fail ", err)
			}
		}
	}
	// log.Println("MEM STORAGE BATCH", metrics.Metrics)
	if batchSend {
		var metric api.Metrics
		resp, err := JSONSendMetrics(
			fmt.Sprintf(updatesAddress, endpoint),
			contentEnc, singKey, metric, metrics)
		if err != nil {
			log.Println(err)
		}
		JSONdecode(resp, batchSend)
	}
}

func iterDB(nojson, batchSend bool, dbconstr, endpoint, contentEnc, singKey string) {
	var metric api.Metrics
	var metrics []api.Metrics

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
		switch {
		case batchSend:
			metrics = append(metrics,
				api.Metrics{MType: "gauge", ID: metric.ID, Value: metric.Value})
		default:
			retrybuilder := func() func() error {
				return func() error {
					err := CommonSendGag(nojson, endpoint, contentEnc, singKey, metric.ID, *metric.Value)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				log.Fatal("send gauge mertic fail ", err)
			}
		}
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
		switch {
		case batchSend:
			metrics = append(metrics,
				api.Metrics{MType: "counter", ID: metric.ID, Delta: metric.Delta})
		default:
			retrybuilder := func() func() error {
				return func() error {
					err := CommonSendCnt(nojson, endpoint, contentEnc, singKey, metric.ID, *metric.Delta)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err := utils.RetrayFunc(retrybuilder)
			if err != nil {
				log.Fatal("send counter mertic fail ", err)
			}
		}
	}
	// log.Println("DATA BASE BATCH", metrics.Metrics)
	if batchSend {
		var metric api.Metrics
		resp, err := JSONSendMetrics(
			fmt.Sprintf(updatesAddress, endpoint),
			contentEnc, singKey, metric, metrics)
		if err != nil {
			log.Println(err)
		}
		JSONdecode(resp, batchSend)
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
}

// func iterFile(nojson bool, fileStoragePath, endpoint, contentEnc string) {
// 	metric := api.Metrics{}
// 	consumer, err := files.NewConsumer(fileStoragePath)
// 	if err != nil {
// 		log.Fatal(err, " can't create consumer in if")
// 	}

// 	scanner := consumer.Scanner
// 	for scanner.Scan() {
// 		// преобразуем данные из JSON-представления в структуру
// 		err := json.Unmarshal(scanner.Bytes(), &metric)
// 		if err != nil {
// 			log.Printf("can't unmarshal string %v", err)
// 		}
// 		if metric.MType == "gauge" {
// 			CommonSendGag(nojson, endpoint, contentEnc, metric.ID, *metric.Value)
// 		}
// 		if metric.MType == "counter" {
// 			CommonSendCnt(nojson, endpoint, contentEnc, metric.ID, *metric.Delta)
// 		}
// 	}
// }

func main() {
	var endpoint string
	var contentEnc string
	var dbconstring string
	var fileStoragePath string
	var singkeystr string
	var nojson bool
	var pInterv int
	var rInterv int
	var err error
	storageSelecter := "MEMORY"
	ctx := context.TODO()
	// устанвливаем для отображения даты и времени в логах
	log.SetFlags(log.Ldate | log.Ltime)

	// storefiledfl := "agentmetrics.json"

	// workDir := utils.WorkingDir()
	// if !utils.ChkFileExist(workDir + saveMetricsDefaultPath) {
	// 	log.Fatal(err)
	// }

	// опредаляем флаги
	pflag.StringVarP(&endpoint, "endpoint", "a", addressServer, "Used to set the address and port to connect server.")
	pflag.StringVarP(&contentEnc, "contentenc", "c", api.Gz, "Used to set content encoding to connect server.")
	// pflag.StringVarP(&fileStoragePath, "filepath", "f", storefiledfl, "Used to set file path to save metrics.")
	pflag.StringVarP(&dbconstring, "dbconstring", "d", "", "Used to set file path to save metrics.")
	pflag.StringVarP(&singkeystr, "singkeystr", "k", "", "Used to set key for calc hash.")
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

	singkeystrTmp := os.Getenv("KEY")
	if len(singkeystrTmp) != 0 {
		singkeystr = singkeystrTmp
	}

	// if fileStoragePath != storefiledfl && len(fileStoragePath) != 0 {
	// 	// storefiledfl = fmt.Sprintf("%s/%s", workDir, fileStoragePath)
	// 	fileStoragePath = fmt.Sprintf("%s/%s", workDir, fileStoragePath)
	// 	storageSelecter = "FILE"
	// }

	// fileStoragePathTMP := os.Getenv("FILE_STORAGE_PATH")
	// if len(fileStoragePathTMP) != 0 {
	// 	// fileStoragePath = fmt.Sprintf("%s/%s", workDir, fileStoragePathTMP)
	// 	fileStoragePath = fileStoragePathTMP
	// 	storageSelecter = "FILE"
	// }

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

	batchSend := ChkUpdates(endpoint)

	pollTik := time.NewTicker(time.Duration(pInterv) * time.Second)
	reportTik := time.NewTicker(time.Duration(rInterv) * time.Second)
	counter := 0

	for {
		select {
		case <-pollTik.C:
			counter += 1

			retrybuilder := func() func() error {
				return func() error {
					err := CollectMetrics(storage, fileStoragePath, dbconstring, storageSelecter, counter)
					if err != nil {
						log.Println(err)
					}
					return err
				}
			}
			err = utils.RetrayFunc(retrybuilder)
			if err != nil {
				log.Fatal("can`t update metrics ", err)
			}

		case <-reportTik.C:
			if storageSelecter == "MEMORY" {
				iterMemStorage(storage, nojson, batchSend, endpoint, contentEnc, singkeystr)
			}
			if storageSelecter == "DATABASE" {
				iterDB(nojson, batchSend, dbconstring, endpoint, contentEnc, singkeystr)
			}
			// if storageSelecter == "FILE" {
			// 	iterFile(nojson, fileStoragePath, endpoint, contentEnc)
			// }
		}
	}
}
