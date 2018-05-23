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
}

func runContainer(w http.ResponseWriter, r *http.Request) (error, interface{}) {
	var runContainer rest.RunContainerJSON

	err := rest.ReceiveJSONFromHttp(w, r, &runContainer)
	if err == nil {
		log.Debugf("<-- RUN Image: %s", runContainer.ContainerImageKey)

		err := thisNode.Scheduler().Deploy(runContainer.ContainerImageKey, runContainer.Arguments,
			runContainer.CPUs, runContainer.RAM)
		return err, nil
	}
	return err, nil
}
