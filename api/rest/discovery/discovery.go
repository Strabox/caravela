package discovery

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	"net/http"
)

var nodeDiscoveryAPI Discovery = nil

func Init(router *mux.Router, nodeDiscovery Discovery) {
	nodeDiscoveryAPI = nodeDiscovery
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(createOffer)).Methods(http.MethodPost)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(refreshOffer)).Methods(http.MethodPatch)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(removeOffer)).Methods(http.MethodDelete)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(getOffers)).Methods(http.MethodGet)
	router.Handle(rest.DiscoveryNeighborOfferBaseEndpoint, rest.AppHandler(neighborOffers)).Methods(http.MethodPatch)
}

func createOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var createOfferMsg rest.CreateOfferMsg

	err := rest.ReceiveJSONFromHttp(w, r, &createOfferMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- CREATE OFFER To: %s, ID: %d, Amt: %d, Res: <%d,%d>, From: %s",
		createOfferMsg.ToNode.GUID, createOfferMsg.Offer.ID, createOfferMsg.Offer.Amount,
		createOfferMsg.Offer.Resources.CPUs, createOfferMsg.Offer.Resources.RAM, createOfferMsg.FromNode.IP)

	nodeDiscoveryAPI.CreateOffer(&createOfferMsg.FromNode, &createOfferMsg.ToNode, &createOfferMsg.Offer)
	return nil, nil
}

func refreshOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var offerRefreshMsg rest.RefreshOfferMsg

	err := rest.ReceiveJSONFromHttp(w, r, &offerRefreshMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- REFRESH OFFER ID: %d, From: %s", offerRefreshMsg.Offer.ID,
		offerRefreshMsg.FromTrader.GUID)

	res := nodeDiscoveryAPI.RefreshOffer(&offerRefreshMsg.FromTrader, &offerRefreshMsg.Offer)
	return rest.RefreshOfferResponseMsg{Refreshed: res}, nil
}

func removeOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var offerRemoveMsg rest.OfferRemoveMsg

	err := rest.ReceiveJSONFromHttp(w, r, &offerRemoveMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- REMOVE OFFER To: %s, ID: %d, From: %s", offerRemoveMsg.ToTrader.GUID,
		offerRemoveMsg.Offer.ID, offerRemoveMsg.FromSupplier.IP)

	nodeDiscoveryAPI.RemoveOffer(&offerRemoveMsg.FromSupplier, &offerRemoveMsg.ToTrader, &offerRemoveMsg.Offer)
	return nil, nil
}

func getOffers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var getOffersMsg rest.GetOffersMsg

	err := rest.ReceiveJSONFromHttp(w, r, &getOffersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- GET OFFERS To: %s", getOffersMsg.ToTrader.GUID)

	return nodeDiscoveryAPI.GetOffers(&getOffersMsg.FromNode, &getOffersMsg.ToTrader, getOffersMsg.Relay), nil
}

func neighborOffers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var neighborOffersMsg rest.NeighborOffersMsg

	err := rest.ReceiveJSONFromHttp(w, r, &neighborOffersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- NEIGHBOR OFFERS To: %s, TraderOffering: <%s;%s>",
		neighborOffersMsg.ToNeighbor.GUID, neighborOffersMsg.NeighborOffering.IP, neighborOffersMsg.NeighborOffering.GUID)

	nodeDiscoveryAPI.AdvertiseOffersNeighbor(&neighborOffersMsg.FromNeighbor, &neighborOffersMsg.ToNeighbor,
		&neighborOffersMsg.NeighborOffering)

	return nil, nil
}
