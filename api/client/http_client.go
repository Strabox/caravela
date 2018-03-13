package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/strabox/caravela/api/rest"
	"net/http"
	"time"
)

const HTTP_CONTENT_TYPE = "application/json"

type HttpClient struct {
	httpClient *http.Client
	apiPort    int
}

func NewHttpClient(apiPort int) *HttpClient {
	res := &HttpClient{}
	res.apiPort = apiPort

	transport := &http.Transport{
		MaxIdleConns: 10,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	res.httpClient = client

	return res
}

func (client *HttpClient) Offer(destIP string, destGuid string, suppIP string, offerID int, amount int) error {
	var offer rest.OfferJSON
	offer.Amount = amount
	offer.DestGuid = destGuid
	offer.SuppIP = suppIP
	offer.OfferID = offerID

	url := fmt.Sprintf("http://%s:%d%s", destIP, client.apiPort, rest.DISCOVERY_BASE_ENDPOINT+rest.DISCOVERY_OFFER_ENDPOINT)

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(offer)

	_, err := client.httpClient.Post(url, HTTP_CONTENT_TYPE, buffer)
	if err == nil {
		fmt.Println("WOOOOT RESPONSE")

		return nil
	} else {
		fmt.Println("[Client] Error: ", err)
		return err
	}
}
