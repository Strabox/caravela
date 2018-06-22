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

func createOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var createOffer rest.CreateOfferMessage

	discovery := thisNode.Discovery()

	err := rest.ReceiveJSONFromHttp(w, r, &createOffer)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- CREATE OFFER FromSuppIP: %s, OfferID: %d, Amount: %d, Resources: <%d,%d>, ToTraderGUID: %s",
		createOffer.FromSupplierIP, createOffer.OfferID, createOffer.Amount, createOffer.CPUs,
		createOffer.RAM, createOffer.ToTraderGUID)

	discovery.CreateOffer(createOffer.FromSupplierGUID, createOffer.FromSupplierIP, createOffer.ToTraderGUID,
		createOffer.OfferID, createOffer.Amount, createOffer.CPUs, createOffer.RAM)
	return nil, nil
}

func refreshOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var offerRefresh rest.RefreshOfferMessage

	discovery := thisNode.Discovery()

	err := rest.ReceiveJSONFromHttp(w, r, &offerRefresh)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- REFRESH OFFER OfferID: %d, FromTraderGUID: %s", offerRefresh.OfferID,
		offerRefresh.FromTraderGUID)

	res := discovery.RefreshOffer(offerRefresh.OfferID, offerRefresh.FromTraderGUID)
	refreshOfferResponseJSON := rest.RefreshOfferResponseMessage{Refreshed: res}
	return refreshOfferResponseJSON, nil
}

func removeOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var offerRemove rest.OfferRemoveMessage

	discovery := thisNode.Discovery()

	err := rest.ReceiveJSONFromHttp(w, r, &offerRemove)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- REMOVE OFFER FromSuppIP: %s, OfferID: %d, ToTraderGUID: %s", offerRemove.FromSupplierIP,
		offerRemove.OfferID, offerRemove.FromSupplierIP, offerRemove.ToTraderGUID)

	discovery.RemoveOffer(offerRemove.FromSupplierIP, offerRemove.FromSupplierGUID,
		offerRemove.ToTraderGUID, offerRemove.OfferID)
	return nil, nil
}

func getOffers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var getOffersJSON rest.GetOffersMessage

	err := rest.ReceiveJSONFromHttp(w, r, &getOffersJSON)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- GET OFFERS ToTraderGUID: %s", getOffersJSON.ToTraderGUID)

	offers := thisNode.Discovery().GetOffers(getOffersJSON.ToTraderGUID)

	var offersJSON []rest.OfferJSON = nil
	offersJSON = make([]rest.OfferJSON, len(offers))
	for index, offer := range offers {
		offersJSON[index].ID = offer.ID
		offersJSON[index].SupplierIP = offer.SupplierIP
	}

	return rest.OffersListMessage{Offers: offersJSON}, nil
}
