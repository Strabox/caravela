package discovery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	nodeAPI "github.com/strabox/caravela/node/api"
	"log"
	"net/http"
)

var thisNode nodeAPI.Node = nil

func InitializeAPI(router *mux.Router, selfNode nodeAPI.Node) {
	thisNode = selfNode
	router.HandleFunc(rest.DISCOVERY_BASE_ENDPOINT+rest.DISCOVERY_OFFER_ENDPOINT, offer).Methods(http.MethodPost)
	router.HandleFunc(rest.DISCOVERY_BASE_ENDPOINT+rest.DISCOVERY_REFERSH_OFFER_ENDPOINT, refreshOffer).Methods(http.MethodGet)
}

func offer(w http.ResponseWriter, r *http.Request) {
	var offer rest.OfferJSON

	//discovery := thisNode.Discovery()

	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(&offer)
		if err == nil {
			log.Println("IT ARRIVED")
			// TODO Update Internals
			http.Error(w, "", http.StatusOK)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

}

func refreshOffer(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("REST WORKING")
}
