package user

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	"net/http"
)

var userNodeAPI User = nil

func Init(router *mux.Router, userNode User) {
	userNodeAPI = userNode
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

	err = userNodeAPI.SubmitContainers(runContainerMsg.ContainerImageKey, runContainerMsg.PortMappings,
		runContainerMsg.Arguments, runContainerMsg.CPUs, runContainerMsg.RAM)
	return nil, err
}

func stopContainers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var err error
	var stopContainersMsg rest.StopContainersMessage

	err = rest.ReceiveJSONFromHttp(w, r, &stopContainersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- STOP Containers: %v", stopContainersMsg.ContainersIDs)

	err = userNodeAPI.StopContainers(stopContainersMsg.ContainersIDs)
	return nil, err
}

func listContainers(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	log.Infof("<-- LIST Containers")

	return userNodeAPI.ListContainers(), nil
}

func exit(_ http.ResponseWriter, _ *http.Request) (interface{}, error) {
	log.Infof("<-- EXITING CARAVELA")

	userNodeAPI.Stop()
	return nil, nil
}
