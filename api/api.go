package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/parvit/qpep/server"
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
		key := fmt.Sprintf(server.QUIC_CONN, addr)
		counter = server.Statistics.Get(key)
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
	dataAddr := strings.Split(r.RemoteAddr, ":")
	port, _ := strconv.ParseInt(dataAddr[1], 10, 64)

	data, err := json.Marshal(EchoResponse{
		Address: dataAddr[0],
		Port:    int(port),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
