package api

import "github.com/julienschmidt/httprouter"

type APIRouter struct {
	handler *httprouter.Router
}

type EchoResponse struct {
	Address string `json:"Address"`
	Port    int    `json:"Port"`
}

type StatusReponse struct {
	LastCheck         string `json:"LastCheck"`
	ConnectionCounter int    `json:"ConnectionCounter"`
}
