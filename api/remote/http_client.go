package remote

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"net/http"
)

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

	log.Infof("--> CREATE OFFER FromSuppIP: %s, OfferID: %d, Amount: %d, Resources: <%d,%d>, ToTraderIP: %s, ToTraderGUID: %s",
		fromSupplierIP, offerID, amount, cpus, ram, toTraderIP, toTraderGUID)

	offerJSON := rest.CreateOfferMessage{FromSupplierIP: fromSupplierIP, FromSupplierGUID: fromSupplierGUID,
		ToTraderGUID: toTraderGUID, OfferID: offerID, Amount: amount, CPUs: cpus, RAM: ram}

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, offerJSON, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *HttpClient) RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (error, bool) {
	log.Infof("--> REFRESH OFFER FromTraderGUID: %s, OfferID: %d, ToSupplierIP: %s",
		fromTraderGUID, offerID, toSupplierIP)

	offerRefreshJSON := rest.RefreshOfferMessage{FromTraderGUID: fromTraderGUID, OfferID: offerID}
	var refreshOfferResponseJSON rest.RefreshOfferResponseMessage

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPatch, offerRefreshJSON,
		&refreshOfferResponseJSON)
	if err != nil {
		return NewRemoteClientError(err), false
	}

	return nil, refreshOfferResponseJSON.Refreshed
}

func (client *HttpClient) RemoveOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string,
	toTraderGUID string, offerID int64) error {

	log.Infof("--> REMOVE OFFER FromSuppIP: %s, FromSuppGUID: %s, OfferID: %d, ToTraderIP: %s, ToTraderGUID: %s",
		fromSupplierIP, fromSupplierGUID, offerID, toTraderIP, toTraderGUID)

	offerRemoveJSON := rest.OfferRemoveMessage{FromSupplierIP: fromSupplierIP, FromSupplierGUID: fromSupplierGUID,
		ToTraderGUID: toTraderGUID, OfferID: offerID}

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodDelete, offerRemoveJSON, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	return nil
}

func (client *HttpClient) GetOffers(toTraderIP string, toTraderGUID string) (error, []Offer) {
	log.Infof("--> GET OFFERS ToTraderIP: %s, ToTraderGUID: %s", toTraderIP, toTraderGUID)

	getOffersJSON := rest.GetOffersMessage{ToTraderGUID: toTraderGUID}
	var offersJSON rest.OffersListMessage

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, getOffersJSON, &offersJSON)
	if err != nil {
		return NewRemoteClientError(err), nil
	}

	if httpCode == http.StatusOK {
		if offersJSON.Offers != nil {
			res := make([]Offer, len(offersJSON.Offers))
			for i, v := range offersJSON.Offers {
				res[i].ID = v.ID
				res[i].SupplierIP = v.SupplierIP
			}
			return nil, res
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}
}

func (client *HttpClient) LaunchContainer(toSupplierIP string, fromBuyerIP string, offerID int64,
	containerImageKey string, containerArgs []string, cpus int, ram int) error {

	log.Infof("--> LAUNCH FromBuyerIP: %s, OfferID: %d, Image: %s, Args: %v, Resources: <%d,%d>, ToSuppIP: %s",
		fromBuyerIP, offerID, containerImageKey, containerArgs, cpus, ram, toSupplierIP)

	launchContainerJSON := rest.LaunchContainerMessage{FromBuyerIP: fromBuyerIP, OfferID: offerID,
		ContainerImageKey: containerImageKey, ContainerArgs: containerArgs, CPUs: cpus, RAM: ram}

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.SchedulerContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, launchContainerJSON, nil)
	if err != nil {
		return NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return nil
	} else {
		return NewRemoteClientError(fmt.Errorf("impossible launch container"))
	}
}

func (client *HttpClient) ObtainConfiguration(systemsNodeIP string) (*configuration.Configuration, error) {
	log.Infof("--> OBTAIN CONFIGS ToNodeIP: %s", systemsNodeIP)

	var systemsNodeConfiguration configuration.Configuration

	url := rest.BuildHttpURL(false, systemsNodeIP, client.config.APIPort(), rest.ConfigurationBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, nil, &systemsNodeConfiguration)
	if err != nil {
		return nil, NewRemoteClientError(err)
	}

	if httpCode == http.StatusOK {
		return &systemsNodeConfiguration, nil
	} else {
		return nil, NewRemoteClientError(fmt.Errorf("impossible obtain node's configurations"))
	}
}
