package api

import (
	"encoding/json"
	"fmt"
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
	var counter float64 = -1.0
	addr := ps.ByName("addr")

	if len(addr) > 0 {
		counter = Statistics.GetCounter(PERF_CONN, addr)
	}

	data, err := json.Marshal(StatusReponse{
		LastCheck:         time.Now().Format(time.RFC3339Nano),
		ConnectionCounter: int(counter),
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
		Address:       dataAddr[0],
		Port:          port,
		ServerVersion: shared.Version(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)

	Statistics.SetState(INFO_PLATFORM, r.Header.Get("User-Agent"))
	Statistics.SetState(INFO_UPDATE, time.Now().Format(time.RFC1123Z))
}

// path /versions
func apiVersions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	server := "N/A"
	client := "N/A"
	if strings.Contains(r.URL.String(), API_PREFIX_SERVER) {
		server = shared.Version()
	} else {
		server = Statistics.GetState(INFO_OTHER_VERSION)
		client = shared.Version()
	}

	data, err := json.Marshal(VersionsResponse{
		Server: server,
		Client: client,
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

// path /statistics/info, /statistics/info/:addr
func apiStatisticsInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reqAddress := ps.ByName("addr")

	lastUpdate := ""
	address := shared.QuicConfiguration.ListenIP
	platform := runtime.GOOS
	if len(reqAddress) > 0 {
		address = reqAddress
		platform = Statistics.GetState(INFO_PLATFORM)
		lastUpdate = Statistics.GetState(INFO_UPDATE)
	} else {
		lastUpdate = time.Now().Format(time.RFC1123Z)
	}

	info := StatsInfoReponse{}
	info.Data = make([]StatsInfoRow, 0, 3)
	info.Data = append(info.Data, StatsInfoRow{
		ID:        1,
		Attribute: "Address",
		Value:     address,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        2,
		Attribute: "Last Update",
		Value:     lastUpdate,
		Name:      INFO_UPDATE,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        3,
		Attribute: "Platform",
		Value:     platform,
		Name:      INFO_PLATFORM,
	})

	data, err := json.Marshal(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// path /statistics/data , /statistics/data/:addr
func apiStatisticsData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reqAddress := ps.ByName("addr")

	currConnections := Statistics.GetCounter(PERF_CONN, reqAddress)
	if currConnections == -1.0 {
		currConnections = 0.0
	}

	upSpeed := Statistics.GetCounter(PERF_UP_SPEED, reqAddress)
	dwSpeed := Statistics.GetCounter(PERF_DW_SPEED, reqAddress)
	upTotal := Statistics.GetCounter(PERF_UP_TOTAL, reqAddress)
	dwTotal := Statistics.GetCounter(PERF_DW_TOTAL, reqAddress)
	if upSpeed == -1.0 {
		upSpeed = 0.0
	}
	if dwSpeed == -1.0 {
		dwSpeed = 0.0
	}
	if upTotal == -1.0 {
		upTotal = 0.0
	}
	if dwTotal == -1.0 {
		dwTotal = 0.0
	}

	info := StatsInfoReponse{}
	info.Data = make([]StatsInfoRow, 0, 5)
	info.Data = append(info.Data, StatsInfoRow{
		ID:        1,
		Attribute: "Current Connections",
		Value:     strconv.Itoa(int(currConnections)),
		Name:      PERF_CONN,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        2,
		Attribute: "Current Upload Speed",
		Value:     fmt.Sprintf("%.2f", upSpeed),
		Name:      PERF_UP_SPEED,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        3,
		Attribute: "Current Download Speed",
		Value:     fmt.Sprintf("%.2f", dwSpeed),
		Name:      PERF_DW_SPEED,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        4,
		Attribute: "Total Uploaded Bytes",
		Value:     fmt.Sprintf("%.2f", upTotal),
		Name:      PERF_UP_TOTAL,
	})
	info.Data = append(info.Data, StatsInfoRow{
		ID:        5,
		Attribute: "Total Downloaded Bytes",
		Value:     fmt.Sprintf("%.2f", dwTotal),
		Name:      PERF_DW_TOTAL,
	})

	data, err := json.Marshal(info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
