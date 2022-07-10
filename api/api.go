package api

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/parvit/qpep/shared"
)

func RunAPIServer(ctx context.Context) {
	listenAddr := shared.QuicConfiguration.ListenIP + ":" + strconv.Itoa(shared.QuicConfiguration.GatewayAPIPort)
	log.Printf("Opening API Server on: %s", listenAddr)

	rtr := NewRouter()
	rtr.registerHandlers()

	srv := NewServer(listenAddr, rtr, ctx)
	log.Println(srv.ListenAndServe())

	log.Println("Closed API Server")
}

func NewServer(addr string, rtr *APIRouter, ctx context.Context) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: rtr.handler,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
}

type APIRouter struct {
	handler *httprouter.Router
}

func NewRouter() *APIRouter {
	rtr := httprouter.New()
	rtr.RedirectTrailingSlash = true
	rtr.RedirectFixedPath = false

	return &APIRouter{
		handler: rtr,
	}
}

func apiHeaders(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next(w, r)
	})
}

func (r *APIRouter) registerHandlers() {
	r.handler.PanicHandler = func(w http.ResponseWriter, r *http.Request, i interface{}) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	r.handler.HandlerFunc(http.MethodGet, "/api/v1/status", apiHeaders(apiV1Status))
}

func apiV1Status(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(StatusReponse{
		LastCheck: time.Now().Format(time.RFC3339Nano),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

type StatusReponse struct {
	LastCheck string `json:"LastCheck"`
}
