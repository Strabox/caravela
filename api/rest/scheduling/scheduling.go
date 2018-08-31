package scheduling

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest/containers"
	"github.com/strabox/caravela/api/rest/util"
	"net/http"
)

var nodeSchedulingAPI Scheduling = nil

func Init(router *mux.Router, nodeScheduling Scheduling) {
	nodeSchedulingAPI = nodeScheduling
	router.Handle(containers.BaseEndpoint, util.AppHandler(launchContainer)).Methods(http.MethodPost)
}

func launchContainer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var err error
	var launchContainerMsg util.LaunchContainerMsg

	err = util.ReceiveJSONFromHttp(w, req, &launchContainerMsg)
	if err != nil {
		return nil, err
	}
	for i, contConfig := range launchContainerMsg.ContainersConfigs {
		log.Infof("<-- LAUNCH [%d] From: %s, ID: %d, Img: %s, PortMaps: %v, Args: %v, Res: <<%d;%d>;%d>",
			i, launchContainerMsg.FromBuyer.IP, launchContainerMsg.Offer.ID, contConfig.ImageKey,
			contConfig.PortMappings, contConfig.Args, contConfig.Resources.CPUClass, contConfig.Resources.CPUs, contConfig.Resources.RAM)
	}

	containersStatus, err := nodeSchedulingAPI.LaunchContainers(req.Context(), &launchContainerMsg.FromBuyer,
		&launchContainerMsg.Offer, launchContainerMsg.ContainersConfigs)
	if err != nil {
		return nil, err
	}

	return containersStatus, err
}
