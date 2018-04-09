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
	router.HandleFunc(rest.DiscoveryOfferBaseEndpoint, createOffer).Methods(http.MethodPost)
	router.HandleFunc(rest.DiscoveryOfferBaseEndpoint, refreshOffer).Methods(http.MethodPatch)
	router.HandleFunc(rest.DiscoveryOfferBaseEndpoint, removeOffer).Methods(http.MethodDelete)
	router.HandleFunc(rest.DiscoveryOfferBaseEndpoint, getOffers).Methods(http.MethodGet)
}

func createOffer(w http.ResponseWriter, r *http.Request) {
	var createOffer rest.CreateOfferJSON

	discovery := thisNode.Discovery()

	if rest.ReceiveJSONFromHttp(w, r, &createOffer) {
		discovery.CreateOffer(createOffer.FromSupplierGUID, createOffer.FromSupplierIP, createOffer.ToTraderGUID,
			createOffer.OfferID, createOffer.Amount, createOffer.CPUs, createOffer.RAM)
		w.WriteHeader(http.StatusOK)
	}
}

func refreshOffer(w http.ResponseWriter, r *http.Request) {
	var offerRefresh rest.RefreshOfferJSON

	discovery := thisNode.Discovery()

	if rest.ReceiveJSONFromHttp(w, r, &offerRefresh) {
		res := discovery.RefreshOffer(offerRefresh.OfferID, offerRefresh.FromTraderGUID)
		if res {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func removeOffer(w http.ResponseWriter, r *http.Request) {
	var offerRemove rest.OfferRemoveJSON

	//discovery := thisNode.Discovery()

	if rest.ReceiveJSONFromHttp(w, r, &offerRemove) {
		// TODO
	}
}

func getOffers(w http.ResponseWriter, r *http.Request) {
	var getOffersJSON rest.GetOffersJSON

	if rest.ReceiveJSONFromHttp(w, r, &getOffersJSON) {
		offers := thisNode.Discovery().GetOffers(getOffersJSON.ToTraderGUID)
		var offersJSON []rest.OfferJSON = nil
		offersJSON = make([]rest.OfferJSON, len(offers))

		for i, o := range offers {
			offersJSON[i].ID = o.ID
			offersJSON[i].SupplierIP = o.SupplierIP
		}

		w.WriteHeader(http.StatusOK)
		w.Write(rest.ToJSONBytes(rest.OffersListJSON{offersJSON}))
	}
}
