package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	"log"
	"net/http"
	"time"
)

const TCPMaxIdleConnections = 10

// Our HTTP body is always a JSON
const HTTPContentType = "application/json"
const HTTPRequestTimeout = 5 * time.Second

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

func (client *HttpClient) CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string, toTraderGUID string,
	offerID int, amount int, cpus int, ram int) *ClientError {

	var offer rest.OfferJSON
	offer.FromSupplierIP = fromSupplierIP
	offer.FromSupplierGUID = fromSupplierGUID
	offer.ToTraderGUID = toTraderGUID
	offer.OfferID = offerID
	offer.Amount = amount
	offer.CPUs = cpus
	offer.RAM = ram

	url := fmt.Sprintf("http://%s:%d%s", toTraderIP, client.config.APIPort(), rest.DiscoveryBaseEndpoint+rest.DiscoveryOfferEndpoint)
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(offer)

	_, err := client.httpClient.Post(url, HTTPContentType, buffer)
	if err == nil {
		log.Println("[Client] Offer received")
		return nil
	} else {
		log.Println("[Client] Offer error: ", err)
		return NewClientError(UNKNOWN)
	}
}

func (client *HttpClient) RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int, responseChan chan<- OfferRefreshResponse) {
	var offerRefresh rest.OfferRefreshJSON
	offerRefresh.OfferID = offerID
	offerRefresh.FromTraderGUID = fromTraderGUID

	url := fmt.Sprintf("http://%s:%d%s", toSupplierIP, client.config.APIPort(), rest.DiscoveryBaseEndpoint+rest.DiscoveryRefreshOfferEndpoint)
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(offerRefresh)

	httpResp, err := client.httpClient.Post(url, HTTPContentType, buffer)
	response := OfferRefreshResponse{toSupplierIP, offerID, false}

	if err == nil && httpResp.StatusCode == http.StatusOK {
		response.Success = true
		responseChan <- response
	} else {
		response.Success = false
		responseChan <- response
	}
}
