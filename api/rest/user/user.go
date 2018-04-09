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
	router.HandleFunc(rest.UserContainerBaseEndpoint, runContainer).Methods(http.MethodPost)
}

func runContainer(w http.ResponseWriter, r *http.Request) {
	var runContainer rest.RunContainerJSON

	if rest.ReceiveJSONFromHttp(w, r, &runContainer) {
		log.Debug(runContainer)
		thisNode.Scheduler().Deploy(runContainer.ContainerImage, runContainer.Arguments, runContainer.CPUs, runContainer.RAM)
	}
}
