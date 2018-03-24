package discovery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	nodeAPI "github.com/strabox/caravela/node/api"
	"net/http"
)

var thisNode nodeAPI.Node = nil

func InitializeAPI(router *mux.Router, selfNode nodeAPI.Node) {
	thisNode = selfNode
	router.HandleFunc(rest.DiscoveryBaseEndpoint+rest.DiscoveryOfferEndpoint, offer).Methods(http.MethodPost)
	router.HandleFunc(rest.DiscoveryBaseEndpoint+rest.DiscoveryRefreshOfferEndpoint, refreshOffer).Methods(http.MethodGet)
}

func offer(w http.ResponseWriter, r *http.Request) {
	var offer rest.OfferJSON

	discovery := thisNode.Discovery()

	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(&offer)
		if err == nil {
			discovery.CreateOffer(offer.FromSupplierGUID, offer.FromSupplierIP, offer.ToTraderGUID, offer.OfferID,
				offer.Amount, offer.CPUs, offer.RAM)
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
	var offerRefresh rest.OfferRefreshJSON

	discovery := thisNode.Discovery()

	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(&offerRefresh)
		if err == nil {
			res := discovery.RefreshOffer(offerRefresh.OfferID, offerRefresh.FromTraderGUID)
			if res {
				http.Error(w, "", http.StatusOK)
			} else {
				http.Error(w, "", http.StatusBadRequest)
			}
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
