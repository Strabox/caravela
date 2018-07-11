package configuration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	"net/http"
)

var nodeConfigurationAPI Configurations = nil

func Init(router *mux.Router, nodeConfiguration Configurations) {
	nodeConfigurationAPI = nodeConfiguration
	router.Handle(rest.ConfigurationBaseEndpoint, rest.AppHandler(obtainConfiguration)).Methods(http.MethodGet)
}

func obtainConfiguration(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	log.Infof("<-- OBTAIN CONFIGS")
	return nodeConfigurationAPI.Configuration(), nil
}
