package discovery

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
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(createOffer)).Methods(http.MethodPost)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(refreshOffer)).Methods(http.MethodPatch)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(removeOffer)).Methods(http.MethodDelete)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(getOffers)).Methods(http.MethodGet)
}

func createOffer(w http.ResponseWriter, r *http.Request) (error, interface{}) {
	var createOffer rest.CreateOfferJSON

	discovery := thisNode.Discovery()

	err := rest.ReceiveJSONFromHttp(w, r, &createOffer)
	if err == nil {
		log.Debugf("<-- CREATE OFFER SuppIP: %s, Resources: <%d,%d>", createOffer.FromSupplierIP, createOffer.CPUs,
			createOffer.RAM)

		discovery.CreateOffer(createOffer.FromSupplierGUID, createOffer.FromSupplierIP, createOffer.ToTraderGUID,
			createOffer.OfferID, createOffer.Amount, createOffer.CPUs, createOffer.RAM)
	}
	return err, nil
}

func refreshOffer(w http.ResponseWriter, r *http.Request) (error, interface{}) {
	var offerRefresh rest.RefreshOfferJSON

	discovery := thisNode.Discovery()

	err := rest.ReceiveJSONFromHttp(w, r, &offerRefresh)
	if err == nil {
		log.Debugf("<-- REFRESH OFFER OfferID: %d, FromTrader: %s", offerRefresh.OfferID,
			offerRefresh.FromTraderGUID)

		res := discovery.RefreshOffer(offerRefresh.OfferID, offerRefresh.FromTraderGUID)
		refreshOfferResponseJSON := rest.RefreshOfferResponseJSON{Refreshed: res}
		return nil, refreshOfferResponseJSON
	}
	return err, nil
}

func removeOffer(w http.ResponseWriter, r *http.Request) (error, interface{}) {
	var offerRemove rest.OfferRemoveJSON

	discovery := thisNode.Discovery()

	err := rest.ReceiveJSONFromHttp(w, r, &offerRemove)
	if err == nil {
		log.Debugf("<-- REMOVE OFFER OfferID: %d, FromSupplier: %s", offerRemove.OfferID,
			offerRemove.FromSupplierIP)

		discovery.RemoveOffer(offerRemove.FromSupplierIP, offerRemove.FromSupplierGUID,
			offerRemove.ToTraderGUID, offerRemove.OfferID)

		return nil, nil
	}
	return err, nil
}

func getOffers(w http.ResponseWriter, r *http.Request) (error, interface{}) {
	var getOffersJSON rest.GetOffersJSON

	err := rest.ReceiveJSONFromHttp(w, r, &getOffersJSON)
	if err == nil {
		log.Debugf("<-- GET OFFERS Trader: %s", getOffersJSON.ToTraderGUID)

		offers := thisNode.Discovery().GetOffers(getOffersJSON.ToTraderGUID)
		var offersJSON []rest.OfferJSON = nil
		offersJSON = make([]rest.OfferJSON, len(offers))

		for i, o := range offers {
			offersJSON[i].ID = o.ID
			offersJSON[i].SupplierIP = o.SupplierIP
		}

		return nil, rest.OffersListJSON{Offers: offersJSON}
	}
	return err, nil
}
