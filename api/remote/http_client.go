package remote

import (
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"net/http"
	"time"
)

const TCPMaxIdleConnections = 10 // TODO: Put in configuration struct?
// Our HTTP body is always a JSON
const HTTPContentType = "application/json" // TODO: Put in configuration struct?
const HTTPRequestTimeout = 2 * time.Second // TODO: Put in configuration struct?

type HttpClient struct {
	httpClient *http.Client
	config     *configuration.Configuration
}

func NewHttpClient(config *configuration.Configuration) *HttpClient {
	res := &HttpClient{}
	res.config = config
	transport := &http.Transport{
		MaxIdleConns: TCPMaxIdleConnections,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   HTTPRequestTimeout,
	}
	res.httpClient = client
	return res
}

func (client *HttpClient) CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string,
	toTraderGUID string, offerID int64, amount int, cpus int, ram int) *Error {

	offerJSON := rest.CreateOfferJSON{fromSupplierIP, fromSupplierGUID,
		toTraderGUID, offerID, amount, cpus, ram}

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, _ := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPost, offerJSON, nil)
	if err == nil {
		return nil
	} else {
		return NewRemoteClientError(err)
	}
}

func (client *HttpClient) RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (*Error, bool) {
	offerRefreshJSON := rest.RefreshOfferJSON{fromTraderGUID, offerID}

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.DiscoveryOfferBaseEndpoint)

	err, httpCode := rest.DoHttpRequestJSON(client.httpClient, url, http.MethodPatch, offerRefreshJSON, nil)
	if err == nil {
		if httpCode == http.StatusOK {
			return nil, true
		} else {
			return nil, false
		}
	} else {
		return NewRemoteClientError(err), false
	}
}

func (client *HttpClient) RemoveOffer(fromSupplierIP string, fromSupplierGUID, toTraderIP string,
	toTraderGUID string, offerID int64) *Error {

	offerRemoveJSON := rest.OfferRemoveJSON{fromSupplierIP, fromSupplierGUID,
		toTraderGUID, offerID}

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
