package user

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	nodeAPI "github.com/strabox/caravela/node/api"
	"net/http"
)

var thisNode nodeAPI.Node = nil

func Initialize(router *mux.Router, selfNode nodeAPI.Node) {
	thisNode = selfNode
	router.Handle(rest.UserContainerBaseEndpoint, rest.AppHandler(runContainer)).Methods(http.MethodPost)
	router.Handle(rest.UserContainerBaseEndpoint, rest.AppHandler(stopContainers)).Methods(http.MethodDelete)
	router.Handle(rest.UserContainerBaseEndpoint, rest.AppHandler(listContainers)).Methods(http.MethodGet)
	router.Handle(rest.UserExitEndpoint, rest.AppHandler(exit)).Methods(http.MethodGet)
}

func runContainer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var err error
	var runContainer rest.RunContainerMessage

	err = rest.ReceiveJSONFromHttp(w, r, &runContainer)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- RUN Image: %s, Args: %v, PortMappings: %v, CPUs: %d, RAM: %d",
		runContainer.ContainerImageKey, runContainer.Arguments, runContainer.PortMappings, runContainer.CPUs,
		runContainer.RAM)

	err = thisNode.Scheduler().Run(runContainer.ContainerImageKey, runContainer.Arguments,
		runContainer.CPUs, runContainer.RAM)
	return nil, err
}

func stopContainers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var err error
	var stopContainers rest.StopContainersMessage

	err = rest.ReceiveJSONFromHttp(w, r, &stopContainers)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- STOP Containers: %v", stopContainers.ContainersIDs)

	// TODO: Forward the call to node
	return nil, nil
}

func listContainers(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	log.Infof("<-- LIST Containers")

	// TODO: Forward the call to node
	return rest.ContainersList{ContainersStatus: make([]rest.ContainerStatus, 0)}, nil
}

func exit(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	log.Infof("<-- EXIT System")

	thisNode.Stop()
	return nil, nil
}
