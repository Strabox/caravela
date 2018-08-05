package containers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	"net/http"
)

var nodeContainersAPI Containers = nil

func Init(router *mux.Router, nodeContainers Containers) {
	nodeContainersAPI = nodeContainers
	router.Handle(rest.ContainersBaseEndpoint, rest.AppHandler(stopLocalContainer)).Methods(http.MethodDelete)
}

func stopLocalContainer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var stopContainerMsg rest.StopLocalContainerMsg

	err := rest.ReceiveJSONFromHttp(w, req, &stopContainerMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- STOP Local Container ID: %s", stopContainerMsg.ContainerID)

	err = nodeContainersAPI.StopLocalContainer(req.Context(), stopContainerMsg.ContainerID)
	return nil, err
}
