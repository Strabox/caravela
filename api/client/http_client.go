package client

import (
	"net/http"
	"encoding/json"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/api/rest/discovery"
	"time"
	"bytes"
	"fmt"
)

const HTTP_CONTENT_TYPE =  "application/json"

type HttpClient struct {
	httpClient *http.Client
}


func NewHttpClient() *HttpClient {
	res := &HttpClient{}
	
	transport := &http.Transport{
		MaxIdleConns:	10,
	}
	
	client := &http.Client{
		Transport: 	transport,
		Timeout:	15 * time.Second,
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
	
	url := fmt.Sprintf("http://%s:%d%s", destIP, 8000, discovery.DISCOVERY_BASE_ENDPOINT +  discovery.DISCOVERY_OFFER_ENDPOINT)
	
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