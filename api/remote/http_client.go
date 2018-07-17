package remote

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/api"
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

func (client *HttpClient) CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string,
	toTraderGUID string, offerID int64, amount int, cpus int, ram int) error {

	log.Infof("--> CREATE OFFER From: %s, ID: %d, Amt: %d, Res: <%d;%d>, To: <%s;%s>",
		fromSupplierIP, offerID, amount, cpus, ram, toTraderIP, toTraderGUID)

	createOfferMsg := rest.CreateOfferMessage{FromSupplierIP: fromSupplierIP, FromSupplierGUID: fromSupplierGUID,
		ToTraderGUID: toTraderGUID, OfferID: offerID, Amount: amount, CPUs: cpus, RAM: ram}

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, createOfferMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *HttpClient) RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (bool, error) {
	log.Infof("--> REFRESH OFFER From: %s, ID: %d, To: %s",
		fromTraderGUID, offerID, toSupplierIP)

	offerRefreshMsg := rest.RefreshOfferMessage{FromTraderGUID: fromTraderGUID, OfferID: offerID}
	var refreshOfferResp rest.RefreshOfferResponseMessage

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPatch, offerRefreshMsg,
		&refreshOfferResp)
	if err != nil {
		return false, NewRemoteClientError(err)
	}

	return refreshOfferResp.Refreshed, nil
}

func (client *HttpClient) RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string,
	toTraderGUID string, offerID int64) error {

	log.Infof("--> REMOVE OFFER From: <%s;%s>, ID: %d, To: <%s;%s>",
		fromSupplierIP, fromSupplierGUID, offerID, toTraderIP, toTraderGUID)

	offerRemoveMsg := rest.OfferRemoveMessage{FromSupplierIP: fromSupplierIP, FromSupplierGUID: fromSupplierGUID,
		ToTraderGUID: toTraderGUID, OfferID: offerID}

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodDelete, offerRemoveMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *HttpClient) GetOffers(toTraderIP string, toTraderGUID string, relay bool, fromNodeGUID string) ([]api.Offer, error) {
	log.Infof("--> GET OFFERS To: <%s;%s>, Relay: %t, From: %s", toTraderIP, toTraderGUID, relay, fromNodeGUID)

	getOffersMsg := rest.GetOffersMessage{
		ToTraderGUID: toTraderGUID,
		Relay:        relay,
		FromNodeGUID: fromNodeGUID,
	}
	var offersResp rest.OffersListMessage

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	client.httpClient.Timeout = 20 * time.Second // TODO: Hack to avoid early timeouts -> Run container sequence of calls should be assynchronous
	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, getOffersMsg, &offersResp)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		if offersResp.Offers != nil {
			res := make([]api.Offer, len(offersResp.Offers))
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

func (client *HttpClient) AdvertiseOffersNeighbor(toNeighborTraderIP string, toNeighborTraderGUID string,
	fromTraderGUID string, traderOfferingGUID string, traderOfferingIP string) error {

	log.Infof("--> NEIGHBOR OFFERS To: <%s;%s> TraderOffering: <%s;%s>", toNeighborTraderIP, toNeighborTraderGUID,
		traderOfferingIP, traderOfferingGUID)

	neighborOfferMsg := rest.NeighborOffersMessage{
		ToNeighborGUID:       toNeighborTraderGUID,
		FromNeighborGUID:     fromTraderGUID,
		NeighborOfferingIP:   traderOfferingIP,
		NeighborOfferingGUID: traderOfferingGUID,
	}

	url := rest.BuildHttpURL(false, toNeighborTraderIP, client.config.APIPort(), rest.DiscoveryNeighborOfferBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPatch, neighborOfferMsg, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewRemoteClientError(fmt.Errorf("impossible advertise neighbor's offers"))
	}
}

func (client *HttpClient) LaunchContainer(toSupplierIP string, fromBuyerIP string, offerID int64,
	containerImageKey string, portMappings []rest.PortMapping, containerArgs []string, cpus int, ram int) (*rest.ContainerStatus, error) {

	log.Infof("--> LAUNCH From: %s, ID: %d, Img: %s, PortMaps: %v, Args: %v, Res: <%d;%d>, To: %s",
		fromBuyerIP, offerID, containerImageKey, portMappings, containerArgs, cpus, ram, toSupplierIP)

	launchContainerMsg := rest.LaunchContainerMessage{
		FromBuyerIP:       fromBuyerIP,
		OfferID:           offerID,
		ContainerImageKey: containerImageKey,
		PortMappings:      portMappings,
		ContainerArgs:     containerArgs,
		CPUs:              cpus,
		RAM:               ram,
	}

	var contStatusResp rest.ContainerStatus

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.ContainersBaseEndpoint)

	client.httpClient.Timeout = 20 * time.Second // TODO: Hack to avoid early timeouts -> Run container sequence of calls should be assynchronous
	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, launchContainerMsg, &contStatusResp)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return &contStatusResp, nil
	} else {
		return nil, NewRemoteClientError(fmt.Errorf("impossible launch container"))
	}
}

func (client *HttpClient) StopLocalContainer(toSupplierIP string, containerID string) error {
	log.Infof("--> STOP ID: %s, SuppIP: %s", containerID, toSupplierIP)

	stopContainerMsg := rest.StopContainerMessage{
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
		return NewRemoteClientError(fmt.Errorf("impossible stop container"))
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
		return nil, NewRemoteClientError(fmt.Errorf("impossible obtain node's configurations"))
	}
}
