package remote

import (
	"bytes"
	"encoding/json"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"log"
	"net/http"
	"time"
)

const TCPMaxIdleConnections = 10

// Our HTTP body is always a JSON
const HTTPContentType = "application/json"
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

	var offer rest.OfferJSON
	offer.FromSupplierIP = fromSupplierIP
	offer.FromSupplierGUID = fromSupplierGUID
	offer.ToTraderGUID = toTraderGUID
	offer.OfferID = offerID
	offer.Amount = amount
	offer.CPUs = cpus
	offer.RAM = ram

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryBaseEndpoint+
		rest.DiscoveryOfferEndpoint)

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(offer)

	_, err := client.httpClient.Post(url, HTTPContentType, buffer)
	if err == nil {
		return nil
	} else {
		return NewClientError(UNKNOWN)
	}
}

func (client *HttpClient) RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (*Error, bool) {
	var offerRefresh rest.OfferRefreshJSON
	offerRefresh.OfferID = offerID
	offerRefresh.FromTraderGUID = fromTraderGUID

	url := rest.BuildHttpURL(false, toSupplierIP, client.config.APIPort(), rest.DiscoveryBaseEndpoint+
		rest.DiscoveryRefreshOfferEndpoint)

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(offerRefresh)

	httpResp, err := client.httpClient.Post(url, HTTPContentType, buffer)

	if err == nil && httpResp.StatusCode == http.StatusOK {
		return nil, true
	} else {
		if err != nil {
			log.Println(err.Error())
		}
		return NewClientError(UNKNOWN), false
	}
}

func (client *HttpClient) RemoveOffer(fromSupplierIP string, fromSupplierGUID, toTraderIP string,
	toTraderGUID string, offerID int64) *Error {

	var offerRemove rest.OfferRemoveJSON
	offerRemove.FromSupplierIP = fromSupplierIP
	offerRemove.FromSupplierGUID = fromSupplierGUID
	offerRemove.ToTraderGUID = toTraderGUID
	offerRemove.OfferID = offerID

	url := rest.BuildHttpURL(false, toTraderIP, client.config.APIPort(), rest.DiscoveryBaseEndpoint+
		rest.DiscoveryRemoveOfferEndpoint)

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(offerRemove)

	_, err := client.httpClient.Post(url, HTTPContentType, buffer)

	if err == nil {
		return nil
	} else {
		if err != nil {
			log.Println(err.Error())
		}
		return NewClientError(UNKNOWN)
	}
}
