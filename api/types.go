package api

import "github.com/julienschmidt/httprouter"

type APIRouter struct {
	handler    *httprouter.Router
	clientMode bool
}

type EchoResponse struct {
	Address       string `json:"address"`
	Port          int64  `json:"port"`
	ServerVersion string `json:"serverversion"`
}

type VersionsResponse struct {
	Server string `json:"server"`
	Client string `json:"client"`
}

type StatusReponse struct {
	LastCheck         string `json:"lastcheck"`
	ConnectionCounter int    `json:"connection_counter"`
}

type StatsInfoRow struct {
	ID        int    `json:"id"`
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}
type StatsInfoReponse struct {
	Data []StatsInfoRow `json:"data"`
}
