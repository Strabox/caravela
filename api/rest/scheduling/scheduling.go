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

func launchContainer(w http.ResponseWriter, r *http.Request) (error, interface{}) {
	var launchJSON rest.LaunchContainerJSON

	err := rest.ReceiveJSONFromHttp(w, r, &launchJSON)
	if err == nil {
		log.Debugf("<-- LAUNCH From: %s , Offer: %d", launchJSON.FromBuyerIP, launchJSON.OfferID)

		err := thisNode.Scheduler().Launch(launchJSON.FromBuyerIP, launchJSON.OfferID, launchJSON.ContainerImageKey,
			launchJSON.ContainerArgs, launchJSON.CPUs, launchJSON.RAM)
		return err, nil
	}
	return err, nil
}
