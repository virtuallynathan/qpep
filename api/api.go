package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"runtime"
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
		break
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
	data, err := json.Marshal(Statistics.GetHosts())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// path /statistics/info
func apiStatisticsClientInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	info := StatsInfoReponse{}
	info.Data = append(info.Data, StatsInfoRow{
		ID:        1,
		Attribute: "Address",
		Value:     shared.QuicConfiguration.ListenIP,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        2,
		Attribute: "Last Update",
		Value:     time.Now().Format(time.RFC1123Z),
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        3,
		Attribute: "Platform",
		Value:     runtime.GOOS,
	})

	data, err := json.Marshal(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// path /statistics/data
func apiStatisticsClientData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	info := StatsInfoReponse{}
	info.Data = make([]StatsInfoRow, 0)

	data, err := json.Marshal(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
