package api

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
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

func Initialize(config *configuration.Configuration, thisNode nodeAPI.Node) {
	log.Debug(util.LogTag("[API]") + "Initializing CARAVELA API ...")

	router = mux.NewRouter()

	// Endpoint used to know everything about the node (Debug Purposes Only)
	router.HandleFunc(rest.DebugEndpoint, debug).Methods(http.MethodGet)

	// Initialize all the API rest endpoints
	discovery.Initialize(router, thisNode)
	scheduling.Initialize(router, thisNode)
	user.Initialize(router, thisNode)

	// Start listening for HTTP requests
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.APIPort()), router))
}

func debug(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("TODO: Debug Endpoint")
}
