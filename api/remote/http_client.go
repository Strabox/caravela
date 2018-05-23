package remote

import (
	"fmt"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"net/http"
	"time"
)

const HTTPRequestTimeout = 10 * time.Second // TODO: Put in configuration struct?

type HttpClient struct {
	httpClient *http.Client
	config     *configuration.Configuration
}

func NewHttpClient(config *configuration.Configuration) *HttpClient {
	res := &HttpClient{}
	res.config = config

	client := &http.Client{
		Timeout: HTTPRequestTimeout,
	}
	res.httpClient = client
	return res
}

func (client *HttpClient) CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string,
	toTraderGUID string, offerID int64, amount int, cpus int, ram int) *Error {

	offerJSON := rest.CreateOfferJSON{FromSupplierIP: fromSupplierIP, FromSupplierGUID: fromSupplierGUID,
		ToTraderGUID: toTraderGUID, OfferID: offerID, Amount: amount, CPUs: cpus, RAM: ram}

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, offerJSON, nil)
	if err == nil {
		return nil
	} else {
		return NewRemoteClientError(err)
	}
}

func (client *HttpClient) RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (*Error, bool) {
	offerRefreshJSON := rest.RefreshOfferJSON{FromTraderGUID: fromTraderGUID, OfferID: offerID}
	var refreshOfferResponseJSON rest.RefreshOfferResponseJSON

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPatch, offerRefreshJSON,
		&refreshOfferResponseJSON)
	if err == nil {
		return nil, refreshOfferResponseJSON.Refreshed
	} else {
		return NewRemoteClientError(err), false
	}
}

func (client *HttpClient) RemoveOffer(fromSupplierIP string, fromSupplierGUID, toTraderIP string,
	toTraderGUID string, offerID int64) *Error {

	offerRemoveJSON := rest.OfferRemoveJSON{FromSupplierIP: fromSupplierIP, FromSupplierGUID: fromSupplierGUID,
		ToTraderGUID: toTraderGUID, OfferID: offerID}

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodDelete, offerRemoveJSON, nil)
	if err == nil {
		return nil
	} else {
		return NewRemoteClientError(err)
	}
}

func (client *HttpClient) GetOffers(toTraderIP string, toTraderGUID string) (*Error, []Offer) {
	getOffersJSON := rest.GetOffersJSON{ToTraderGUID: toTraderGUID}
	var offersJSON rest.OffersListJSON

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodGet, getOffersJSON, &offersJSON)
	if err == nil {
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
	} else {
		return NewRemoteClientError(err), nil
	}
}

func (client *HttpClient) LaunchContainer(toSupplierIP string, fromBuyerIP string, offerID int64,
	containerImageKey string, containerArgs []string, cpus int, ram int) *Error {

	launchContainerJSON := rest.LaunchContainerJSON{FromBuyerIP: fromBuyerIP, OfferID: offerID,
		ContainerImageKey: containerImageKey, ContainerArgs: containerArgs, CPUs: cpus, RAM: ram}

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.SchedulerContainerBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, launchContainerJSON, nil)
	if err == nil {
		if httpCode == http.StatusOK {
			return nil
		} else {
			return NewRemoteClientError(fmt.Errorf("impossible launch container"))
		}
	} else {
		return NewRemoteClientError(err)
	}
}
