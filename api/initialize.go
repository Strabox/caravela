package api

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	configREST "github.com/strabox/caravela/api/rest/configuration"
	"github.com/strabox/caravela/api/rest/discovery"
	"github.com/strabox/caravela/api/rest/scheduling"
	"github.com/strabox/caravela/api/rest/user"
	"github.com/strabox/caravela/configuration"
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/util"
	"net/http"
)

/*
REST router for the http requests.
*/
var router *mux.Router = nil

func Initialize(config *configuration.Configuration, thisNode nodeAPI.Node) error {
	log.Debug(util.LogTag("[API]") + "Initializing CARAVELA REST API ...")

	router = mux.NewRouter()

	// Initialize all the API rest endpoints
	configREST.Initialize(router, thisNode, config)
	discovery.Initialize(router, thisNode)
	scheduling.Initialize(router, thisNode)
	user.Initialize(router, thisNode)

	// Start listening for HTTP requests
	return http.ListenAndServe(fmt.Sprintf(":%d", config.APIPort()), router)
}
