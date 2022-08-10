package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/parvit/qpep/shared"
)

func formatRequest(r *http.Request) string {
	data, err := httputil.DumpRequest(r, shared.QuicConfiguration.Verbose)
	if err != nil {
		return fmt.Sprintf("REQUEST: %v", err)
	}

	return string(data)
}

// path /status
func apiStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var counter int = -1
	addr := ps.ByName("addr")

	if len(addr) > 0 {
		key := Statistics.AsKey(QUIC_CONN, addr)
		counter = Statistics.Get(key)
	}

	data, err := json.Marshal(StatusReponse{
		LastCheck:         time.Now().Format(time.RFC3339Nano),
		ConnectionCounter: counter,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// path /echo
func apiEcho(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	mappedAddr := r.RemoteAddr
	if !strings.HasPrefix(r.RemoteAddr, "127.") {
		mappedAddr = Statistics.GetMappedAddress(r.RemoteAddr)
		log.Printf("remote: %s / mapped: %s\n", r.RemoteAddr, mappedAddr)
		if len(mappedAddr) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	dataAddr := strings.Split(mappedAddr, ":")
	port := int64(0)

	switch len(dataAddr) {
	default:
		fallthrough
	case 0:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case 1:
		break
	case 2:
		port, _ = strconv.ParseInt(dataAddr[1], 10, 64)
	}

	data, err := json.Marshal(EchoResponse{
		Address: dataAddr[0],
		Port:    port,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// path /statistics/hosts
func apiStatisticsHosts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	info := StatsInfoReponse{}
	hosts := Statistics.GetHosts()

	sort.Strings(hosts)
	for i := 0; i < len(hosts); i++ {
		info.Data = append(info.Data, StatsInfoRow{
			ID:        i + 1,
			Attribute: "Address",
			Value:     hosts[i],
		})
	}

	data, err := json.Marshal(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// path /statistics/info, /statistics/:addr/info
func apiStatisticsInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	info := StatsInfoReponse{}
	reqAddress := ps.ByName("addr")

	tm := time.Now().Format(time.RFC1123Z)
	address := shared.QuicConfiguration.ListenIP
	platform := runtime.GOOS
	if len(reqAddress) > 0 {
		address = reqAddress
		//platform =
	}

	info.Data = append(info.Data, StatsInfoRow{
		ID:        1,
		Attribute: "Address",
		Value:     address,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        2,
		Attribute: "Last Update",
		Value:     tm,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        3,
		Attribute: "Platform",
		Value:     platform,
	})

	data, err := json.Marshal(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

var totalUp = 0.0
var totalDw = 0.0

// path /statistics/data , /statistics/:addr/data
func apiStatisticsData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//reqAddress := ps.ByName("addr")

	var up = rand.Float64() * 1000.0
	totalUp += up
	var dw = rand.Float64() * 1000.0
	totalDw += dw

	info := StatsInfoReponse{}
	info.Data = make([]StatsInfoRow, 0, 32)
	info.Data = append(info.Data, StatsInfoRow{
		ID:        1,
		Attribute: "Current Connections",
		Value:     strconv.Itoa(rand.Intn(20)),
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        2,
		Attribute: "Current Download Speed",
		Value:     fmt.Sprintf("%.2f", dw),
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        3,
		Attribute: "Current Upload Speed",
		Value:     fmt.Sprintf("%.2f", up),
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        4,
		Attribute: "Total Downloaded Bytes",
		Value:     fmt.Sprintf("%.2f", totalDw),
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        5,
		Attribute: "Total Uploaded Bytes",
		Value:     fmt.Sprintf("%.2f", totalUp),
	})

	data, err := json.Marshal(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
