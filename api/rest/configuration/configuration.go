package configuration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	nodeAPI "github.com/strabox/caravela/node/api"
	"net/http"
)

var thisNode nodeAPI.Node = nil
var thisNodeConfigs *configuration.Configuration

func Initialize(router *mux.Router, selfNode nodeAPI.Node, thisConfigs *configuration.Configuration) {
	thisNode = selfNode
	thisNodeConfigs = thisConfigs
	router.Handle(rest.ConfigurationBaseEndpoint, rest.AppHandler(obtainConfiguration)).Methods(http.MethodGet)
}

func obtainConfiguration(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	log.Infof("<-- OBTAIN CONFIGS")
	return thisNodeConfigs, nil
}
