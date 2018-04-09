package scheduling

import (
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	nodeAPI "github.com/strabox/caravela/node/api"
	"net/http"
)

var thisNode nodeAPI.Node = nil

func Initialize(router *mux.Router, selfNode nodeAPI.Node) {
	thisNode = selfNode
	router.HandleFunc(rest.SchedulerContainerBaseEndpoint, launchContainer).Methods(http.MethodPost)
}

func launchContainer(w http.ResponseWriter, r *http.Request) {

}
