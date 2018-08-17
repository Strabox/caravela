package configuration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest/util"
	"net/http"
)

const BaseEndpoint = "/configuration"

var nodeConfigurationAPI Configurations = nil

func Init(router *mux.Router, nodeConfiguration Configurations) {
	nodeConfigurationAPI = nodeConfiguration
	router.Handle(BaseEndpoint, util.AppHandler(obtainConfiguration)).Methods(http.MethodGet)
}

func obtainConfiguration(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	log.Infof("<-- OBTAIN CONFIGS")
	return nodeConfigurationAPI.Configuration(req.Context()), nil
}
