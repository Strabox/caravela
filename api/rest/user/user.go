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
	var runContainerMsg rest.RunContainerMessage

	err = rest.ReceiveJSONFromHttp(w, r, &runContainerMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- RUN Img: %s, Args: %v, PortMaps: %v, Res: <%d;%d>",
		runContainerMsg.ContainerImageKey, runContainerMsg.Arguments, runContainerMsg.PortMappings,
		runContainerMsg.CPUs, runContainerMsg.RAM)

	err = thisNode.Scheduler().Run(runContainerMsg.ContainerImageKey, runContainerMsg.PortMappings,
		runContainerMsg.Arguments, runContainerMsg.CPUs, runContainerMsg.RAM)
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
	log.Infof("<-- EXITING CARAVELA")

	thisNode.Stop()
	return nil, nil
}
