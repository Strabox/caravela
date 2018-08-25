package remote

import (
	"context"
	"errors"
	log "github.com/Sirupsen/logrus"
	configREST "github.com/strabox/caravela/api/rest/configuration"
	"github.com/strabox/caravela/api/rest/containers"
	"github.com/strabox/caravela/api/rest/discovery"
	"github.com/strabox/caravela/api/rest/util"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/discovery/offering/partitions"
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
	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	log.Infof("--> CREATE OFFER From: %s, ID: %d, Amt: %d, Res: <%d;%d>, To: <%s;%s>",
		fromNode.IP, offer.ID, offer.Amount, offer.Resources.CPUs, offer.Resources.RAM, toNode.IP, toNode.GUID[0:12])

	createOfferMsg := util.CreateOfferMsg{
		FromNode: *fromNode,
		ToNode:   *toNode,
		Offer:    *offer,
	}

	url := util.BuildHttpURL(false, toNode.IP, client.config.APIPort(), discovery.OfferBaseEndpoint)

	err, _ := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPost, createOfferMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *Client) RefreshOffer(ctx context.Context, fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	log.Infof("--> REFRESH OFFER From: %s, ID: %d, To: %s", fromTrader.GUID[0:12], offer.ID, toSupp.IP)

	refreshOfferMsg := util.RefreshOfferMsg{
		FromTrader: *fromTrader,
		Offer:      *offer,
	}
	var refreshOfferResp util.RefreshOfferResponseMsg

	url := util.BuildHttpURL(false, toSupp.IP, client.config.APIPort(), discovery.OfferBaseEndpoint)

	err, _ := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPatch, refreshOfferMsg,
		&refreshOfferResp)
	if err != nil {
		return false, NewRemoteClientError(err)
	}

	return refreshOfferResp.Refreshed, nil
}

func (client *Client) UpdateOffer(ctx context.Context, fromSupplier, toTrader *types.Node, offer *types.Offer) error {
	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	log.Infof("--> UPDATE OFFER From: %s, ID: %d, To: %s", fromSupplier.IP, offer.ID, toTrader.GUID[0:12])

	updateOfferMsg := util.UpdateOfferMsg{
		FromSupplier: *fromSupplier,
		ToTrader:     *toTrader,
		Offer:        *offer,
	}
	var refreshOfferResp util.RefreshOfferResponseMsg

	url := util.BuildHttpURL(false, toTrader.IP, client.config.APIPort(), discovery.OfferBaseEndpoint)

	err, _ := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPut, updateOfferMsg,
		&refreshOfferResp)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *Client) RemoveOffer(ctx context.Context, fromSupp, toTrader *types.Node, offer *types.Offer) error {
	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	log.Infof("--> REMOVE OFFER From: <%s;%s>, ID: %d, To: <%s;%s>",
		fromSupp.IP, fromSupp.GUID, offer.ID, toTrader.IP, toTrader.GUID[0:12])

	offerRemoveMsg := util.OfferRemoveMsg{
		FromSupplier: *fromSupp,
		ToTrader:     *toTrader,
		Offer:        *offer}

	url := util.BuildHttpURL(false, toTrader.IP, client.config.APIPort(), discovery.OfferBaseEndpoint)

	err, _ := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodDelete, offerRemoveMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *Client) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	log.Infof("--> GET OFFERS To: <%s;%s>, Relay: %t, From: %s", toTrader.IP, toTrader.GUID[0:12], relay, fromNode.GUID)

	getOffersMsg := util.GetOffersMsg{
		FromNode: *fromNode,
		ToTrader: *toTrader,
		Relay:    relay,
	}
	var offers []types.AvailableOffer

	url := util.BuildHttpURL(false, toTrader.IP, client.config.APIPort(), discovery.OfferBaseEndpoint)

	client.httpClient.Timeout = 20 * time.Second // TODO: Hack to avoid early timeouts -> Run container sequence of calls should be assynchronous
	err, httpCode := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodGet, getOffersMsg, &offers)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		if offers != nil {
			return offers, nil
		}
		return nil, nil
	}

	return nil, nil
}

func (client *Client) AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) error {
	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	log.Infof("--> NEIGHBOR OFFERS To: <%s;%s> TraderOffering: <%s;%s>", toNeighborTrader.IP, toNeighborTrader.GUID[0:12],
		traderOffering.IP, traderOffering.GUID[0:12])

	neighborOfferMsg := util.NeighborOffersMsg{
		FromNeighbor:     *fromTrader,
		ToNeighbor:       *toNeighborTrader,
		NeighborOffering: *traderOffering,
	}

	url := util.BuildHttpURL(false, toNeighborTrader.IP, client.config.APIPort(), discovery.NeighborOfferBaseEndpoint)

	err, httpCode := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPatch, neighborOfferMsg, nil)
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

	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	for i, contConfig := range containersConfigs {
		log.Infof("--> LAUNCH [%d] From: %s, ID: %d, Img: %s, PortMaps: %v, Args: %v, Res: <%d;%d>, To: %s",
			i, fromBuyer.IP, offer.ID, contConfig.ImageKey, contConfig.PortMappings, contConfig.Args,
			contConfig.Resources.CPUs, contConfig.Resources.RAM, toSupplier.IP)
	}

	launchContainerMsg := util.LaunchContainerMsg{
		FromBuyer:         *fromBuyer,
		Offer:             *offer,
		ContainersConfigs: containersConfigs,
	}

	var contStatusResp []types.ContainerStatus

	url := util.BuildHttpURL(false, toSupplier.IP, client.config.APIPort(), containers.BaseEndpoint)

	client.httpClient.Timeout = 20 * time.Second // TODO: Hack to avoid early timeouts -> Run container sequence of calls should be assynchronous
	err, httpCode := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodPost, launchContainerMsg, &contStatusResp)
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
	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	log.Infof("--> STOP ID: %s, SuppIP: %s", containerID, toSupplier.IP)

	stopLocalContainerMsg := util.StopLocalContainerMsg{
		ContainerID: containerID,
	}

	url := util.BuildHttpURL(false, toSupplier.IP, client.config.APIPort(), containers.BaseEndpoint)

	err, httpCode := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodDelete, stopLocalContainerMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewRemoteClientError(errors.New("impossible stop container"))
	}
}

func (client *Client) ObtainConfiguration(ctx context.Context, systemsNode *types.Node) (*configuration.Configuration, error) {
	ctx = context.WithValue(context.Background(), types.PartitionsStateKey, partitions.GlobalState.PartitionsState())
	log.Infof("--> OBTAIN CONFIGS To: %s", systemsNode.IP)
	var systemsNodeConfigsResp configuration.Configuration

	url := util.BuildHttpURL(false, systemsNode.IP, client.config.APIPort(), configREST.BaseEndpoint)

	err, httpCode := util.DoHttpRequestJSON(ctx, client.httpClient, url, http.MethodGet, nil, &systemsNodeConfigsResp)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return &systemsNodeConfigsResp, nil
	} else {
		return nil, NewRemoteClientError(errors.New("impossible obtain node's configurations"))
	}
}
