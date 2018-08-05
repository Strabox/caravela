package remote

import (
	"context"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"net/http"
	"time"
)

// Client is used to contact the REST API of other nodes.
type Client struct {
	httpClient *http.Client
	config     *configuration.Configuration
}

func NewClient(config *configuration.Configuration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: config.APITimeout(),
		},
		config: config,
	}
}

func (client *Client) CreateOffer(ctx context.Context, fromNode, toNode *types.Node, offer *types.Offer) error {

	log.Infof("--> CREATE OFFER From: %s, ID: %d, Amt: %d, Res: <%d;%d>, To: <%s;%s>",
		fromNode.IP, offer.ID, offer.Amount, offer.Resources.CPUs, offer.Resources.RAM, toNode.IP, toNode.GUID)

	createOfferMsg := rest.CreateOfferMsg{
		FromNode: *fromNode,
		ToNode:   *toNode,
		Offer:    *offer,
	}

	url := rest.BuildHttpURL(false, toNode.IP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPost, createOfferMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *Client) RefreshOffer(ctx context.Context, fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	log.Infof("--> REFRESH OFFER From: %s, ID: %d, To: %s",
		fromTrader.GUID, offer.ID, toSupp.IP)

	offerRefreshMsg := rest.RefreshOfferMsg{
		FromTrader: *fromTrader,
		Offer:      *offer,
	}
	var refreshOfferResp rest.RefreshOfferResponseMsg

	url := rest.BuildHttpURL(false, toSupp.IP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPatch, offerRefreshMsg,
		&refreshOfferResp)
	if err != nil {
		return false, NewRemoteClientError(err)
	}

	return refreshOfferResp.Refreshed, nil
}

func (client *Client) RemoveOffer(ctx context.Context, fromSupp, toTrader *types.Node, offer *types.Offer) error {
	log.Infof("--> REMOVE OFFER From: <%s;%s>, ID: %d, To: <%s;%s>",
		fromSupp.IP, fromSupp.GUID, offer.ID, toTrader.IP, toTrader.GUID)

	offerRemoveMsg := rest.OfferRemoveMsg{
		FromSupplier: *fromSupp,
		ToTrader:     *toTrader,
		Offer:        *offer}

	url := rest.BuildHttpURL(false, toTrader.IP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodDelete, offerRemoveMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *Client) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	log.Infof("--> GET OFFERS To: <%s;%s>, Relay: %t, From: %s", toTrader.IP, toTrader.GUID, relay, fromNode.GUID)

	getOffersMsg := rest.GetOffersMsg{
		FromNode: *fromNode,
		ToTrader: *toTrader,
		Relay:    relay,
	}
	var offers []types.AvailableOffer

	url := rest.BuildHttpURL(false, toTrader.IP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	client.httpClient.Timeout = 20 * time.Second // TODO: Hack to avoid early timeouts -> Run container sequence of calls should be assynchronous
	err, httpCode := rest.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodGet, getOffersMsg, &offers)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		if offers != nil {
			res := make([]types.AvailableOffer, len(offers))
			for i, v := range offers {
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

func (client *Client) AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) error {

	log.Infof("--> NEIGHBOR OFFERS To: <%s;%s> TraderOffering: <%s;%s>", toNeighborTrader.IP, toNeighborTrader.GUID,
		traderOffering.IP, traderOffering.GUID)

	neighborOfferMsg := rest.NeighborOffersMsg{
		FromNeighbor:     *fromTrader,
		ToNeighbor:       *toNeighborTrader,
		NeighborOffering: *traderOffering,
	}

	url := rest.BuildHttpURL(false, toNeighborTrader.IP, client.config.APIPort(), rest.DiscoveryNeighborOfferBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPatch, neighborOfferMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewRemoteClientError(errors.New("impossible advertise neighbor's offers"))
	}
}

func (client *Client) LaunchContainer(ctx context.Context, fromBuyer, toSupplier *types.Node, offer *types.Offer,
	containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {

	for i, contConfig := range containersConfigs {
		log.Infof("--> LAUNCH [%d] From: %s, ID: %d, Img: %s, PortMaps: %v, Args: %v, Res: <%d;%d>, To: %s",
			i, fromBuyer.IP, offer.ID, contConfig.ImageKey, contConfig.PortMappings, contConfig.Args,
			contConfig.Resources.CPUs, contConfig.Resources.RAM, toSupplier.IP)
	}

	launchContainerMsg := rest.LaunchContainerMsg{
		FromBuyer:         *fromBuyer,
		Offer:             *offer,
		ContainersConfigs: containersConfigs,
	}

	var contStatusResp []types.ContainerStatus

	url := rest.BuildHttpURL(false, toSupplier.IP, client.config.APIPort(), rest.ContainersBaseEndpoint)

	client.httpClient.Timeout = 20 * time.Second // TODO: Hack to avoid early timeouts -> Run container sequence of calls should be assynchronous
	err, httpCode := rest.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPost, launchContainerMsg, &contStatusResp)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return contStatusResp, nil
	} else {
		return nil, NewRemoteClientError(errors.New("impossible launch container"))
	}
}

func (client *Client) StopLocalContainer(ctx context.Context, toSupplier *types.Node, containerID string) error {
	log.Infof("--> STOP ID: %s, SuppIP: %s", containerID, toSupplier.IP)

	stopLocalContainerMsg := rest.StopLocalContainerMsg{
		ContainerID: containerID,
	}

	url := rest.BuildHttpURL(false, toSupplier.IP, client.config.APIPort(), rest.ContainersBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodDelete, stopLocalContainerMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewRemoteClientError(errors.New("impossible stop container"))
	}
}

func (client *Client) ObtainConfiguration(c context.Context, systemsNode *types.Node) (*configuration.Configuration, error) {
	log.Infof("--> OBTAIN CONFIGS To: %s", systemsNode.IP)

	var systemsNodeConfigsResp configuration.Configuration

	url := rest.BuildHttpURL(false, systemsNode.IP, client.config.APIPort(), rest.ConfigurationBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(c, client.httpClient, url, http.MethodGet, nil, &systemsNodeConfigsResp)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return &systemsNodeConfigsResp, nil
	} else {
		return nil, NewRemoteClientError(errors.New("impossible obtain node's configurations"))
	}
}
