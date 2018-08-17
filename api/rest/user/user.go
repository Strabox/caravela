package user

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest/util"
	"github.com/strabox/caravela/api/types"
	"net/http"
)

const baseEndpoint = "/user"
const ContainerBaseEndpoint = baseEndpoint + "/container"
const ExitEndpoint = baseEndpoint + "/exit"

var userNodeAPI User = nil

func Init(router *mux.Router, userNode User) {
	userNodeAPI = userNode
	router.Handle(ContainerBaseEndpoint, util.AppHandler(runContainer)).Methods(http.MethodPost)
	router.Handle(ContainerBaseEndpoint, util.AppHandler(stopContainers)).Methods(http.MethodDelete)
	router.Handle(ContainerBaseEndpoint, util.AppHandler(listContainers)).Methods(http.MethodGet)
	router.Handle(ExitEndpoint, util.AppHandler(exit)).Methods(http.MethodGet)
}

func runContainer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var err error
	var runContainerConfigs []types.ContainerConfig

	err = util.ReceiveJSONFromHttp(w, req, &runContainerConfigs)
	if err != nil {
		return nil, err
	}
	for i, containerConfig := range runContainerConfigs {
		log.Infof("<-- RUN [%d] Img: %s, Args: %v, PortMaps: %v, Res: <%d;%d;%d>, GrpPolicy: %d",
			i, containerConfig.ImageKey, containerConfig.Args, containerConfig.PortMappings,
			containerConfig.Resources.CPUPower, containerConfig.Resources.CPUs, containerConfig.Resources.RAM,
			containerConfig.GroupPolicy)
	}

	return nil, userNodeAPI.SubmitContainers(req.Context(), runContainerConfigs)
}

func stopContainers(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var err error
	var stopContainersIDs []string

	err = util.ReceiveJSONFromHttp(w, req, &stopContainersIDs)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- STOP Containers: %v", stopContainersIDs)

	return nil, userNodeAPI.StopContainers(req.Context(), stopContainersIDs)
}

func listContainers(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	log.Infof("<-- LIST Containers")

	return userNodeAPI.ListContainers(req.Context()), nil
}

func exit(_ http.ResponseWriter, req *http.Request) (interface{}, error) {
	log.Infof("<-- EXITING CARAVELA")

	userNodeAPI.Stop(req.Context())
	return nil, nil
}
