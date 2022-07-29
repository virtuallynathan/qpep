package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
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

func apiStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var counter int = -1
	addr := ps.ByName("addr")

	if len(addr) > 0 {
		key := fmt.Sprintf(QUIC_CONN, addr)
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

func apiEcho(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	mappedAddr := Statistics.GetMappedAddress(r.RemoteAddr)
	log.Printf("remote: %s / mapped: %s\n", r.RemoteAddr, mappedAddr)
	if len(mappedAddr) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
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
