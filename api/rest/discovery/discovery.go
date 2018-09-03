package discovery

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest/util"
	"net/http"
)

const baseEndpoint = "/discovery"
const OfferBaseEndpoint = baseEndpoint + "/offer"
const NeighborOfferBaseEndpoint = baseEndpoint + "/neighbor/offer"

var nodeDiscoveryAPI Discovery = nil

func Init(router *mux.Router, nodeDiscovery Discovery) {
	nodeDiscoveryAPI = nodeDiscovery
	router.Handle(OfferBaseEndpoint, util.AppHandler(createOffer)).Methods(http.MethodPost)
	router.Handle(OfferBaseEndpoint, util.AppHandler(refreshOffer)).Methods(http.MethodPatch)
	router.Handle(OfferBaseEndpoint, util.AppHandler(updateOffer)).Methods(http.MethodPut)
	router.Handle(OfferBaseEndpoint, util.AppHandler(removeOffer)).Methods(http.MethodDelete)
	router.Handle(OfferBaseEndpoint, util.AppHandler(getOffers)).Methods(http.MethodGet)
	router.Handle(NeighborOfferBaseEndpoint, util.AppHandler(neighborOffers)).Methods(http.MethodPatch)
}

func createOffer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var createOfferMsg util.CreateOfferMsg

	err := util.ReceiveJSONFromHttp(w, req, &createOfferMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- CREATE OFFER To: %s, ID: %d, Amt: %d, Res: <%d,%d>, From: %s",
		createOfferMsg.ToNode.GUID[0:12], createOfferMsg.Offer.ID, createOfferMsg.Offer.Amount,
		createOfferMsg.Offer.FreeResources.CPUs, createOfferMsg.Offer.FreeResources.RAM, createOfferMsg.FromNode.IP)

	nodeDiscoveryAPI.CreateOffer(req.Context(), &createOfferMsg.FromNode, &createOfferMsg.ToNode, &createOfferMsg.Offer)
	return nil, nil
}

func refreshOffer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var refreshOfferMsg util.RefreshOfferMsg

	err := util.ReceiveJSONFromHttp(w, req, &refreshOfferMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- REFRESH OFFER ID: %d, From: %s", refreshOfferMsg.Offer.ID,
		refreshOfferMsg.FromTrader.GUID[0:12])

	res := nodeDiscoveryAPI.RefreshOffer(req.Context(), &refreshOfferMsg.FromTrader, &refreshOfferMsg.Offer)
	return util.RefreshOfferResponseMsg{Refreshed: res}, nil
}

func updateOffer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var updateOfferMsg util.UpdateOfferMsg

	err := util.ReceiveJSONFromHttp(w, req, &updateOfferMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- UPDATE OFFER ID: %d, From: %s, To: %s", updateOfferMsg.Offer.ID,
		updateOfferMsg.FromSupplier.IP, updateOfferMsg.ToTrader.GUID[0:12])

	nodeDiscoveryAPI.UpdateOffer(req.Context(), &updateOfferMsg.FromSupplier, &updateOfferMsg.ToTrader, &updateOfferMsg.Offer)
	return nil, nil
}

func removeOffer(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var offerRemoveMsg util.OfferRemoveMsg

	err := util.ReceiveJSONFromHttp(w, req, &offerRemoveMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- REMOVE OFFER To: %s, ID: %d, From: %s", offerRemoveMsg.ToTrader.GUID[0:12],
		offerRemoveMsg.Offer.ID, offerRemoveMsg.FromSupplier.IP)

	nodeDiscoveryAPI.RemoveOffer(req.Context(), &offerRemoveMsg.FromSupplier, &offerRemoveMsg.ToTrader, &offerRemoveMsg.Offer)
	return nil, nil
}

func getOffers(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var getOffersMsg util.GetOffersMsg

	err := util.ReceiveJSONFromHttp(w, req, &getOffersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- GET OFFERS To: %s", getOffersMsg.ToTrader.GUID[0:12])

	return nodeDiscoveryAPI.GetOffers(req.Context(), &getOffersMsg.FromNode, &getOffersMsg.ToTrader, getOffersMsg.Relay), nil
}

func neighborOffers(w http.ResponseWriter, req *http.Request) (interface{}, error) {
	var neighborOffersMsg util.NeighborOffersMsg

	err := util.ReceiveJSONFromHttp(w, req, &neighborOffersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- NEIGHBOR OFFERS To: %s, TraderOffering: <%s;%s>",
		neighborOffersMsg.ToNeighbor.GUID[0:12], neighborOffersMsg.NeighborOffering.IP, neighborOffersMsg.NeighborOffering.GUID[0:12])

	nodeDiscoveryAPI.AdvertiseOffersNeighbor(req.Context(), &neighborOffersMsg.FromNeighbor, &neighborOffersMsg.ToNeighbor,
		&neighborOffersMsg.NeighborOffering)

	return nil, nil
}
