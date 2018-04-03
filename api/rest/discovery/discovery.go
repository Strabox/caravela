package discovery

import (
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	nodeAPI "github.com/strabox/caravela/node/api"
	"net/http"
)

var thisNode nodeAPI.Node = nil

func Initialize(router *mux.Router, selfNode nodeAPI.Node) {
	thisNode = selfNode
	router.HandleFunc(rest.DiscoveryBaseEndpoint+rest.DiscoveryOfferEndpoint, createOffer).Methods(http.MethodPost)
	router.HandleFunc(rest.DiscoveryBaseEndpoint+rest.DiscoveryRefreshOfferEndpoint, refreshOffer).Methods(http.MethodPost)
	router.HandleFunc(rest.DiscoveryBaseEndpoint+rest.DiscoveryRemoveOfferEndpoint, removeOffer).Methods(http.MethodPost)
}

func createOffer(w http.ResponseWriter, r *http.Request) {
	var createOffer rest.OfferJSON

	discovery := thisNode.Discovery()

	if rest.VerifyAndExtractJson(w, r, &createOffer) {
		discovery.CreateOffer(createOffer.FromSupplierGUID, createOffer.FromSupplierIP, createOffer.ToTraderGUID,
			createOffer.OfferID, createOffer.Amount, createOffer.CPUs, createOffer.RAM)
		http.Error(w, "", http.StatusOK)
	}
}

func refreshOffer(w http.ResponseWriter, r *http.Request) {
	var offerRefresh rest.OfferRefreshJSON

	discovery := thisNode.Discovery()

	if rest.VerifyAndExtractJson(w, r, &offerRefresh) {
		res := discovery.RefreshOffer(offerRefresh.OfferID, offerRefresh.FromTraderGUID)
		if res {
			http.Error(w, "", http.StatusOK)
		} else {
			http.Error(w, "", http.StatusBadRequest)
		}
	}
}

func removeOffer(w http.ResponseWriter, r *http.Request) {
	var offerRemove rest.OfferRemoveJSON

	//discovery := thisNode.Discovery()

	if rest.VerifyAndExtractJson(w, r, &offerRemove) {
		// TODO
	}
}
