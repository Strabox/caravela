package scheduling

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
	router.Handle(rest.SchedulerContainerBaseEndpoint, rest.AppHandler(launchContainer)).Methods(http.MethodPost)
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

	err = thisNode.Scheduler().Launch(launchContainerMsg.FromBuyerIP, launchContainerMsg.OfferID,
		launchContainerMsg.ContainerImageKey, launchContainerMsg.PortMappings, launchContainerMsg.ContainerArgs,
		launchContainerMsg.CPUs, launchContainerMsg.RAM)
	return nil, err
}
