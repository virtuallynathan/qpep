package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/parvit/qpep/shared"
)

const (
	API_ECHO_PATH   string = "/api/v1/echo"
	API_STATUS_PATH string = "/api/v1/status/:addr"
)

func RunServer(ctx context.Context, cancel context.CancelFunc) {
	// update configuration from flags
	host := shared.QuicConfiguration.ListenIP
	apiPort := shared.QuicConfiguration.GatewayAPIPort

	listenAddr := fmt.Sprintf("%s:%d", host, apiPort)
	log.Printf("Opening API Server on: %s", listenAddr)

	rtr := NewRouter()
	rtr.registerHandlers()

	srv := NewServer(listenAddr, rtr, ctx)
	go func() {
		<-ctx.Done()
		if srv != nil {
			srv.Close()
			srv = nil
		}
		cancel()
	}()

	if err := srv.ListenAndServe(); err != nil {
		log.Printf("Error running API server: %v", err)
	}
	srv = nil
	cancel()

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
		log.Printf("%s\n", formatRequest(r))

		w.Header().Add("Content-Type", "application/json")
		next(w, r, ps)
	})
}

type notFoundHandler struct{}

func (n *notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s\n", formatRequest(r))
	w.WriteHeader(http.StatusNotFound)
}

type methodsNotAllowedHandler struct{}

func (n *methodsNotAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s\n", formatRequest(r))
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (r *APIRouter) registerHandlers() {
	r.handler.PanicHandler = func(w http.ResponseWriter, r *http.Request, i interface{}) {
		log.Printf("%s\n", formatRequest(r))
		w.WriteHeader(http.StatusInternalServerError)
	}
	r.handler.NotFound = &notFoundHandler{}
	r.handler.MethodNotAllowed = &methodsNotAllowedHandler{}

	r.handler.HandleMethodNotAllowed = true
	r.handler.GET(API_ECHO_PATH, apiHeaders(apiEcho))
	r.handler.GET(API_STATUS_PATH, apiHeaders(apiStatus))
}
