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

func createOffer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var createOfferMsg rest.CreateOfferMsg

	err := rest.ReceiveJSONFromHttp(w, req, &createOfferMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- CREATE OFFER To: %s, ID: %d, Amt: %d, Res: <%d,%d>, From: %s",
		createOfferMsg.ToNode.GUID[0:12], createOfferMsg.Offer.ID, createOfferMsg.Offer.Amount,
		createOfferMsg.Offer.Resources.CPUs, createOfferMsg.Offer.Resources.RAM, createOfferMsg.FromNode.IP)

	nodeDiscoveryAPI.CreateOffer(req.Context(), &createOfferMsg.FromNode, &createOfferMsg.ToNode, &createOfferMsg.Offer)
	return nil, nil
}

func refreshOffer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var offerRefreshMsg rest.RefreshOfferMsg

	err := rest.ReceiveJSONFromHttp(w, req, &offerRefreshMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- REFRESH OFFER ID: %d, From: %s", offerRefreshMsg.Offer.ID,
		offerRefreshMsg.FromTrader.GUID[0:12])

	res := nodeDiscoveryAPI.RefreshOffer(req.Context(), &offerRefreshMsg.FromTrader, &offerRefreshMsg.Offer)
	return rest.RefreshOfferResponseMsg{Refreshed: res}, nil
}

func removeOffer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var offerRemoveMsg rest.OfferRemoveMsg

	err := rest.ReceiveJSONFromHttp(w, req, &offerRemoveMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- REMOVE OFFER To: %s, ID: %d, From: %s", offerRemoveMsg.ToTrader.GUID[0:12],
		offerRemoveMsg.Offer.ID, offerRemoveMsg.FromSupplier.IP)

	nodeDiscoveryAPI.RemoveOffer(req.Context(), &offerRemoveMsg.FromSupplier, &offerRemoveMsg.ToTrader, &offerRemoveMsg.Offer)
	return nil, nil
}

func getOffers(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var getOffersMsg rest.GetOffersMsg

	err := rest.ReceiveJSONFromHttp(w, req, &getOffersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- GET OFFERS To: %s", getOffersMsg.ToTrader.GUID[0:12])

	return nodeDiscoveryAPI.GetOffers(req.Context(), &getOffersMsg.FromNode, &getOffersMsg.ToTrader, getOffersMsg.Relay), nil
}

func neighborOffers(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var neighborOffersMsg rest.NeighborOffersMsg

	err := rest.ReceiveJSONFromHttp(w, req, &neighborOffersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- NEIGHBOR OFFERS To: %s, TraderOffering: <%s;%s>",
		neighborOffersMsg.ToNeighbor.GUID[0:12], neighborOffersMsg.NeighborOffering.IP, neighborOffersMsg.NeighborOffering.GUID[0:12])

	nodeDiscoveryAPI.AdvertiseOffersNeighbor(req.Context(), &neighborOffersMsg.FromNeighbor, &neighborOffersMsg.ToNeighbor,
		&neighborOffersMsg.NeighborOffering)

	return nil, nil
}
