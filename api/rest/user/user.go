package user

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/api/types"
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
	var runContainerConfigs []types.ContainerConfig

	err = rest.ReceiveJSONFromHttp(w, r, &runContainerConfigs)
	if err != nil {
		return nil, err
	}
	for i, containerConfig := range runContainerConfigs {
		log.Infof("<-- RUN [%d] Img: %s, Args: %v, PortMaps: %v, Res: <%d;%d;%d>, GrpPolicy: %d",
			i, containerConfig.ImageKey, containerConfig.Args, containerConfig.PortMappings,
			containerConfig.Resources.CPUPower, containerConfig.Resources.CPUs, containerConfig.Resources.RAM,
			containerConfig.GroupPolicy)
	}

	return nil, userNodeAPI.SubmitContainers(runContainerConfigs)
}

func stopContainers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var err error
	var stopContainersIDs []string

	err = rest.ReceiveJSONFromHttp(w, r, &stopContainersIDs)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- STOP Containers: %v", stopContainersIDs)

	return nil, userNodeAPI.StopContainers(stopContainersIDs)
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
