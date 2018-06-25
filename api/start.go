package api

import (
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

/* REST router for the HTTP requests. */
var router *mux.Router = nil

func Start(config *configuration.Configuration, thisNode nodeAPI.Node) (*http.Server, error) {
	log.Debug(util.LogTag("[API]") + "Starting REST API ...")

	router = mux.NewRouter()

	// Start all the API rest endpoints and the respective endpoint routing
	configEndpoint.Initialize(router, thisNode, config)
	discovery.Initialize(router, thisNode)
	scheduling.Initialize(router, thisNode)
	user.Initialize(router, thisNode)

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", config.APIPort()), Handler: router}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Infof(util.LogTag("[API]")+" REST API server STOPPED: %s", err)
		}
	}()

	return httpServer, nil
}
