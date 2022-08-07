package api

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/parvit/qpep/shared"
	"github.com/parvit/qpep/webgui"
)

const (
	API_PREFIX_SERVER string = "/api/v1/server"
	API_PREFIX_CLIENT string = "/api/v1/client"

	API_ECHO_PATH        string = "/echo"
	API_STATUS_PATH      string = "/status/:addr"
	API_STATS_HOSTS_PATH string = "/statistics/hosts"
	API_STATS_INFO_PATH  string = "/statistics/info"
	API_STATS_DATA_PATH  string = "/statistics/data"
)

func RunAPIServer(ctx context.Context, clientMode bool) {
	// update configuration from flags
	host := shared.QuicConfiguration.ListenIP
	if clientMode {
		host = "127.0.0.1"
		log.Printf("Ignored listening address for api server in client mode, forced to 127.0.0.1")
	}
	apiPort := shared.QuicConfiguration.GatewayAPIPort

	listenAddr := fmt.Sprintf("%s:%d", host, apiPort)
	log.Printf("Opening API Server on: %s", listenAddr)

	rtr := NewRouter()
	rtr.clientMode = clientMode
	rtr.registerHandlers()
	rtr.registerStaticFiles()

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

func NewRouter() *APIRouter {
	rtr := httprouter.New()
	rtr.RedirectTrailingSlash = true
	rtr.RedirectFixedPath = true

	return &APIRouter{
		handler: rtr,
	}
}

func apiFilter(next httprouter.Handle) httprouter.Handle {
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
	r.handler.RedirectTrailingSlash = false
	r.handler.HandleMethodNotAllowed = true

	// register apis with respective allowed usage
	r.registerAPIMethod("GET", API_ECHO_PATH, apiFilter(apiEcho), true, true)
	r.registerAPIMethod("GET", API_STATUS_PATH, apiFilter(apiStatus), true, false)

	r.registerAPIMethod("GET", API_STATS_HOSTS_PATH, apiFilter(apiStatisticsHosts), true, false)
	r.registerAPIMethod("GET", API_STATS_INFO_PATH, apiFilter(apiStatisticsClientInfo), false, true)
	r.registerAPIMethod("GET", API_STATS_DATA_PATH, apiFilter(apiStatisticsClientData), false, true)
}

func (r *APIRouter) registerAPIMethod(method, path string, handle httprouter.Handle, allowServer, allowClient bool) {
	if !allowServer && !allowClient {
		panic(fmt.Sprintf("Requested registration of api method %s %s for neither server or client usage!", method, path))
	}

	log.Printf("Register API: %s %s (srv:%v cli:%v cli-mode:%v)\n", method, path, allowServer, allowClient, r.clientMode)
	if allowServer && !r.clientMode {
		r.handler.Handle(method, API_PREFIX_SERVER+path, handle)
	} else {
		r.handler.Handle(method, API_PREFIX_SERVER+path, apiForbidden)
	}

	if allowClient && r.clientMode {
		r.handler.Handle(method, API_PREFIX_CLIENT+path, handle)
	} else {
		r.handler.Handle(method, API_PREFIX_CLIENT+path, apiForbidden)
	}
}

func apiForbidden(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Printf("%s\n", formatRequest(r))

	w.WriteHeader(http.StatusForbidden)
}

func (r *APIRouter) registerStaticFiles() {
	r.handler.GET("/", redirectHome)
	r.handler.GET("/index", redirectHome)

	for path, _ := range webgui.FilesList {
		r.handler.GET("/"+path, serveFile)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	urlPath := r.URL.Path[1:]

	typeFile := mime.TypeByExtension(urlPath)
	w.Header().Add("Content-Type", typeFile)

	w.Write(webgui.FilesList[urlPath])
}

func redirectHome(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query := r.URL.Query().Encode()

	if len(query) > 0 {
		http.Redirect(w, r, "/home?"+query, http.StatusPermanentRedirect)
		return
	}
	http.Redirect(w, r, "/home", http.StatusPermanentRedirect)
}
