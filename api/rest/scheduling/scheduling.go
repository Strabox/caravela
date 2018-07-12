package scheduling

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	"net/http"
)

var nodeSchedulingAPI Scheduling = nil

func Init(router *mux.Router, nodeScheduling Scheduling) {
	nodeSchedulingAPI = nodeScheduling
	router.Handle(rest.ContainersBaseEndpoint, rest.AppHandler(launchContainer)).Methods(http.MethodPost)
}

func launchContainer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var err error
	var launchContainerMsg rest.LaunchContainerMessage

	err = rest.ReceiveJSONFromHttp(w, r, &launchContainerMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- LAUNCH From: %s, ID: %d, Img: %s, PortMaps: %v, Args: %v, Res: <%d;%d>",
		launchContainerMsg.FromBuyerIP, launchContainerMsg.OfferID, launchContainerMsg.ContainerImageKey,
		launchContainerMsg.PortMappings, launchContainerMsg.ContainerArgs, launchContainerMsg.CPUs,
		launchContainerMsg.RAM)

	containerID, err := nodeSchedulingAPI.LaunchContainers(launchContainerMsg.FromBuyerIP, launchContainerMsg.OfferID,
		launchContainerMsg.ContainerImageKey, launchContainerMsg.PortMappings, launchContainerMsg.ContainerArgs,
		launchContainerMsg.CPUs, launchContainerMsg.RAM)

	if err != nil {
		return nil, err
	}

	contStatusResp := &rest.ContainerStatus{
		ID: containerID,
	}

	return contStatusResp, err
}
