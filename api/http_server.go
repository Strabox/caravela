package api

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest/configuration"
	"github.com/strabox/caravela/api/rest/containers"
	"github.com/strabox/caravela/api/rest/discovery"
	"github.com/strabox/caravela/api/rest/scheduling"
	"github.com/strabox/caravela/api/rest/user"
	"github.com/strabox/caravela/util"
	"net/http"
)

// HttpServer handles the REST API requests in each node and redirects it to the local Node,
// where is the logic's core.
type HttpServer struct {
	router     *mux.Router
	httpServer *http.Server
}

// NewServer creates a new API HttpServer that receives the requests for the local node.
func NewServer(port int) *HttpServer {
	router := mux.NewRouter()
	return &HttpServer{
		router: router, // HTTP request endpoint router
		httpServer: &http.Server{ // Filled when server is started
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
	}
}

// Start initializes the endpoints and starts the http web server.
func (server *HttpServer) Start(node LocalNode) error {
	log.Debug(util.LogTag("[API]") + "Starting REST API HttpServer ...")

	// Initialize all the API rest endpoints
	configuration.Init(server.router, node)
	containers.Init(server.router, node)
	discovery.Init(server.router, node)
	scheduling.Init(server.router, node)
	user.Init(server.router, node)

	// Starts the http web server
	go func() {
		if err := server.httpServer.ListenAndServe(); err != nil {
			log.Infof(util.LogTag("[API]")+" REST API server STOPPED: %s", err)
		}
	}()

	return nil
}

// Stop the http web server
func (server *HttpServer) Stop() {
	go server.httpServer.Shutdown(context.Background())
}
