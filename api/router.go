package api

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/parvit/qpep/shared"
	"github.com/parvit/qpep/webgui"
	"github.com/rs/cors"
)

const (
	API_PREFIX_SERVER string = "/api/v1/server"
	API_PREFIX_CLIENT string = "/api/v1/client"

	API_ECHO_PATH           string = "/echo"
	API_STATUS_PATH         string = "/status/:addr"
	API_STATS_HOSTS_PATH    string = "/statistics/hosts"
	API_STATS_INFO_PATH     string = "/statistics/info"
	API_STATS_DATA_PATH     string = "/statistics/data"
	API_STATS_INFO_SRV_PATH string = "/statistics/info/:addr"
	API_STATS_DATA_SRV_PATH string = "/statistics/data/:addr"
)

func RunAPIServer(ctx context.Context, localMode bool) {
	// update configuration from flags
	host := shared.QuicConfiguration.ListenIP
	if localMode {
		host = "127.0.0.1"
		log.Printf("Ignored listening address for local api server, forced to 127.0.0.1")
	}
	apiPort := shared.QuicConfiguration.GatewayAPIPort

	listenAddr := fmt.Sprintf("%s:%d", host, apiPort)
	log.Printf("Opening API Server on: %s", listenAddr)

	rtr := NewRouter()
	rtr.clientMode = shared.QuicConfiguration.ClientFlag
	rtr.registerHandlers()
	rtr.registerStaticFiles()

	srv := NewServer(listenAddr, rtr, ctx)
	log.Println(srv.ListenAndServe())

	log.Println("Closed API Server")
}

func NewServer(addr string, rtr *APIRouter, ctx context.Context) *http.Server {
	corsRouterHandler := cors.Default().Handler(rtr.handler)

	return &http.Server{
		Addr:    addr,
		Handler: corsRouterHandler,
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

		// Request API request must accept JSON
		accepts := r.Header.Get(textproto.CanonicalMIMEHeaderKey("accept"))
		if len(accepts) > 0 {
			if !strings.Contains(accepts, "application/json") &&
				!strings.Contains(accepts, "application/*") &&
				!strings.Contains(accepts, "*/*") {

				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		// Request is found for API request
		w.Header().Add("Content-Type", "application/json")
		next(w, r, ps)
	})
}

type notFoundHandler struct{}

func (n *notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s\n", formatRequest(r))

	// Request not found for API request will accept JSON
	accepts := r.Header.Get(textproto.CanonicalMIMEHeaderKey("accept"))
	if len(accepts) > 0 && strings.EqualFold(accepts, "application/json") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Request not found for non-API serves the default page
	serveFile(w, r, nil)
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
	r.registerAPIMethod("GET", API_STATS_INFO_PATH, apiFilter(apiStatisticsInfo), false, true)
	r.registerAPIMethod("GET", API_STATS_DATA_PATH, apiFilter(apiStatisticsData), true, true)
	r.registerAPIMethod("GET", API_STATS_INFO_SRV_PATH, apiFilter(apiStatisticsInfo), true, false)
	r.registerAPIMethod("GET", API_STATS_DATA_SRV_PATH, apiFilter(apiStatisticsData), true, false)
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
	for path, _ := range webgui.FilesList {
		if path == "index.html" {
			continue
		}
		r.handler.GET("/"+path, serveFile)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	urlPath := r.URL.Path[1:]
	if _, ok := webgui.FilesList[urlPath]; !ok {
		urlPath = "index.html"
	}

	var typeFile string
	if len(filepath.Ext(urlPath)) == 0 {
		typeFile = "text/html"
	} else {
		typeFile = mime.TypeByExtension(urlPath)
	}

	w.Header().Add("Content-Type", typeFile)

	w.Write(webgui.FilesList[urlPath])
}
