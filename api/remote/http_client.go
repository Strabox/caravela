package remote

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"net/http"
	"time"
)

// HttpClient is used to contact the REST API of other nodes.
type HttpClient struct {
	httpClient *http.Client
	config     *configuration.Configuration
}

func NewHttpClient(config *configuration.Configuration) *HttpClient {
	return &HttpClient{
		httpClient: &http.Client{
			Timeout: config.APITimeout(),
		},
		config: config,
	}
}

func (client *HttpClient) CreateOffer(fromNode, toNode *types.Node, offer *types.Offer) error {

	log.Infof("--> CREATE OFFER From: %s, ID: %d, Amt: %d, Res: <%d;%d>, To: <%s;%s>",
		fromNode.IP, offer.ID, offer.Amount, offer.Resources.CPUs, offer.Resources.RAM, toNode.IP, toNode.GUID)

	createOfferMsg := rest.CreateOfferMsg{
		FromNode: *fromNode,
		ToNode:   *toNode,
		Offer:    *offer,
	}

	url := rest.BuildHttpURL(false, toNode.IP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, createOfferMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *HttpClient) RefreshOffer(fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	log.Infof("--> REFRESH OFFER From: %s, ID: %d, To: %s",
		fromTrader.GUID, offer.ID, toSupp.IP)

	offerRefreshMsg := rest.RefreshOfferMsg{
		FromTrader: *fromTrader,
		Offer:      *offer,
	}
	var refreshOfferResp rest.RefreshOfferResponseMsg

	url := rest.BuildHttpURL(false, toSupp.IP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPatch, offerRefreshMsg,
		&refreshOfferResp)
	if err != nil {
		return false, NewRemoteClientError(err)
	}

	return refreshOfferResp.Refreshed, nil
}

func (client *HttpClient) RemoveOffer(fromSupp, toTrader *types.Node, offer *types.Offer) error {
	log.Infof("--> REMOVE OFFER From: <%s;%s>, ID: %d, To: <%s;%s>",
		fromSupp.IP, fromSupp.GUID, offer.ID, toTrader.IP, toTrader.GUID)

	offerRemoveMsg := rest.OfferRemoveMsg{
		FromSupplier: *fromSupp,
		ToTrader:     *toTrader,
		Offer:        *offer}

	url := rest.BuildHttpURL(false, toTrader.IP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodDelete, offerRemoveMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *HttpClient) GetOffers(fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	log.Infof("--> GET OFFERS To: <%s;%s>, Relay: %t, From: %s", toTrader.IP, toTrader.GUID, relay, fromNode.GUID)

	getOffersMsg := rest.GetOffersMsg{
		FromNode: *fromNode,
		ToTrader: *toTrader,
		Relay:    relay,
	}
	var offersResp rest.OffersListMsg

	url := rest.BuildHttpURL(false, toTrader.IP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	client.httpClient.Timeout = 20 * time.Second // TODO: Hack to avoid early timeouts -> Run container sequence of calls should be assynchronous
	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, getOffersMsg, &offersResp)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		if offersResp.Offers != nil {
			res := make([]types.AvailableOffer, len(offersResp.Offers))
			for i, v := range offersResp.Offers {
				res[i].ID = v.ID
				res[i].SupplierIP = v.SupplierIP
			}
			return res, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}
}

func (client *HttpClient) AdvertiseOffersNeighbor(fromTrader, toNeighborTrader, traderOffering *types.Node) error {

	log.Infof("--> NEIGHBOR OFFERS To: <%s;%s> TraderOffering: <%s;%s>", toNeighborTrader.IP, toNeighborTrader.GUID,
		traderOffering.IP, traderOffering.GUID)

	neighborOfferMsg := rest.NeighborOffersMsg{
		FromNeighbor:     *fromTrader,
		ToNeighbor:       *toNeighborTrader,
		NeighborOffering: *traderOffering,
	}

	url := rest.BuildHttpURL(false, toNeighborTrader.IP, client.config.APIPort(), rest.DiscoveryNeighborOfferBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPatch, neighborOfferMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewRemoteClientError(errors.New("impossible advertise neighbor's offers"))
	}
}

func (client *HttpClient) LaunchContainer(fromBuyer, toSupplier *types.Node, offer *types.Offer,
	containerConfig *types.ContainerConfig) (*types.ContainerStatus, error) {

	log.Infof("--> LAUNCH From: %s, ID: %d, Img: %s, PortMaps: %v, Args: %v, Res: <%d;%d>, To: %s",
		fromBuyer.IP, offer.ID, containerConfig.ImageKey, containerConfig.PortMappings, containerConfig.Args,
		containerConfig.Resources.CPUs, containerConfig.Resources.RAM, toSupplier.IP)

	launchContainerMsg := rest.LaunchContainerMsg{
		FromBuyer:       *fromBuyer,
		Offer:           *offer,
		ContainerConfig: *containerConfig,
	}

	var contStatusResp types.ContainerStatus

	url := rest.BuildHttpURL(false, toSupplier.IP, client.config.APIPort(), rest.ContainersBaseEndpoint)

	client.httpClient.Timeout = 20 * time.Second // TODO: Hack to avoid early timeouts -> Run container sequence of calls should be assynchronous
	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, launchContainerMsg, &contStatusResp)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return &contStatusResp, nil
	} else {
		return nil, NewRemoteClientError(errors.New("impossible launch container"))
	}
}

func (client *HttpClient) StopLocalContainer(toSupplierIP string, containerID string) error {
	log.Infof("--> STOP ID: %s, SuppIP: %s", containerID, toSupplierIP)

	stopContainerMsg := rest.StopContainerMsg{
		ContainerID: containerID,
	}

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.ContainersBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodDelete, stopContainerMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewRemoteClientError(errors.New("impossible stop container"))
	}
}

func (client *HttpClient) ObtainConfiguration(systemsNodeIP string) (*configuration.Configuration, error) {
	log.Infof("--> OBTAIN CONFIGS To: %s", systemsNodeIP)

	var systemsNodeConfigsResp configuration.Configuration

	url := rest.BuildHttpURL(false, systemsNodeIP, client.config.APIPort(), rest.ConfigurationBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, nil, &systemsNodeConfigsResp)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return &systemsNodeConfigsResp, nil
	} else {
		return nil, NewRemoteClientError(errors.New("impossible obtain node's configurations"))
	}
}
