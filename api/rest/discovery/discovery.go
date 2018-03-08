package discovery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/node/local"
	"github.com/strabox/caravela/api/rest"
	"net/http"
	"fmt"
)

const DISCOVERY_BASE_ENDPOINT = "/discovery"

const DISCOVERY_OFFER_ENDPOINT = "/offer"
const DISCOVERY_REFERSH_OFFER_ENDPOINT = "/refresh"

var thisNode local.LocalNode = nil


func InitializeDiscoveryAPI(router *mux.Router, selfNode local.LocalNode) {
	thisNode = selfNode
	router.HandleFunc(DISCOVERY_BASE_ENDPOINT + DISCOVERY_OFFER_ENDPOINT, offer).Methods(http.MethodPost)
	router.HandleFunc(DISCOVERY_BASE_ENDPOINT + DISCOVERY_REFERSH_OFFER_ENDPOINT, refreshOffer).Methods(http.MethodGet)
}

func offer(w http.ResponseWriter, r *http.Request) {
	var offer rest.OfferJSON
	
	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(&offer)
		if err == nil {
			fmt.Println("IT ARRIVED")
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