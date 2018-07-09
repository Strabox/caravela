package api

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	configEndpoint "github.com/strabox/caravela/api/rest/configuration"
	"github.com/strabox/caravela/api/rest/discovery"
	"github.com/strabox/caravela/api/rest/scheduling"
	"github.com/strabox/caravela/api/rest/user"
	"github.com/strabox/caravela/configuration"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/util"
	"net/http"
)

/*
Represents the server (HTTP server) that handle the API requests in each node.
*/
type HttpServer struct {
	router     *mux.Router
	httpServer *http.Server
}

func NewServer() *HttpServer {
	return &HttpServer{
		router:     mux.NewRouter(), // HTTP request endpoint router
		httpServer: nil,             // Filled when server is started
	}
}

func (server *HttpServer) Start(config *configuration.Configuration, thisNode nodeAPI.Node) error {
	log.Debug(util.LogTag("[API]") + "Starting REST API HttpServer ...")

	// Start all the API rest endpoints and the respective endpoint routing
	configEndpoint.Initialize(server.router, thisNode, config)
	discovery.Initialize(server.router, thisNode)
	scheduling.Initialize(server.router, thisNode)
	user.Initialize(server.router, thisNode)

	server.httpServer = &http.Server{Addr: fmt.Sprintf(":%d", config.APIPort()), Handler: server.router}

	go func() {
		if err := server.httpServer.ListenAndServe(); err != nil {
			log.Infof(util.LogTag("[API]")+" REST API server STOPPED: %s", err)
		}
	}()

	return nil
}

func (server *HttpServer) Stop() {
	go server.httpServer.Shutdown(context.Background())
}
