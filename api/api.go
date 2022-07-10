package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
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
	rtr.RedirectFixedPath = true

	return &APIRouter{
		handler: rtr,
	}
}

func apiHeaders(next httprouter.Handle) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Printf("0 %s\n", formatRequest(r))

		w.Header().Add("Content-Type", "application/json")
		next(w, r, ps)
	})
}

type notFoundHandler struct{}

func (n *notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("1 %s\n", formatRequest(r))
	w.WriteHeader(http.StatusNotFound)
}

type methodsNotAllowedHandler struct{}

func (n *methodsNotAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("2 %s\n", formatRequest(r))
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (r *APIRouter) registerHandlers() {
	r.handler.PanicHandler = func(w http.ResponseWriter, r *http.Request, i interface{}) {
		log.Printf("3 %s\n", formatRequest(r))
		w.WriteHeader(http.StatusInternalServerError)
	}
	r.handler.NotFound = &notFoundHandler{}
	r.handler.MethodNotAllowed = &methodsNotAllowedHandler{}

	r.handler.HandleMethodNotAllowed = true
	r.handler.GET("/api/v1/status/:addr", apiHeaders(apiStatus))
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

type StatusReponse struct {
	LastCheck         string `json:"LastCheck"`
	ConnectionCounter int    `json:"ConnectionCounter"`
}
