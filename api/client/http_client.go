package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/strabox/caravela/api/rest"
	"log"
	"net/http"
	"time"
)

// Our HTTP body is always a JSON
const HTTP_CONTENT_TYPE = "application/json"

const TCP_MAX_IDLE_CONNS = 10
const HTTP_REQUEST_TIMEOUT = 5 * time.Second

type HttpClient struct {
	httpClient *http.Client
	apiPort    int
}

func NewHttpClient(apiPort int) *HttpClient {
	res := &HttpClient{}
	res.apiPort = apiPort

	transport := &http.Transport{
		MaxIdleConns: TCP_MAX_IDLE_CONNS,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   HTTP_REQUEST_TIMEOUT,
	}

	res.httpClient = client

	return res
}

func (client *HttpClient) Offer(destTraderIP string, destTraderGUID string, suppIP string,
	suppGUID string, offerID int, amount int) *ClientError {

	var offer rest.OfferJSON
	offer.TraderDestGUID = destTraderGUID
	offer.SupplierIP = suppIP
	offer.SupplierGUID = suppGUID
	offer.OfferID = offerID
	offer.Amount = amount

	url := fmt.Sprintf("http://%s:%d%s", destTraderIP, client.apiPort, rest.DISCOVERY_BASE_ENDPOINT+rest.DISCOVERY_OFFER_ENDPOINT)

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(offer)

	_, err := client.httpClient.Post(url, HTTP_CONTENT_TYPE, buffer)
	if err == nil {
		log.Println("[Client] Offer received")
		return nil
	} else {
		log.Println("[Client] Offer error: ", err)
		return NewClientError(UNKNOWN)
	}
}

func (client *HttpClient) RefreshOffer(destSupplierIP string, traderGUID string, offerID int) *ClientError {
	// TODO
	return nil
}

func (client *HttpClient) RemoveOffer(destTraderIP string, destTraderGUID string, suppGUID string, offerID int) *ClientError {
	// TODO
	return nil
}
