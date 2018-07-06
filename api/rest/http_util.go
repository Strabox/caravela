package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/util"
	"net/http"
)

/*
Validate the HTTP message content extracting the JSON into a golang structure if necessary
*/
func ReceiveJSONFromHttp(_ http.ResponseWriter, r *http.Request, jsonToFill interface{}) error {
	if r.Body != nil { // Verify if HTTP message body is not empty
		err := json.NewDecoder(r.Body).Decode(jsonToFill)
		if err == nil { // Verify if JSON was decoded with success
			return nil
		} else {
			return err
		}
	} else {
		return fmt.Errorf("empty HTTP body when JSON was expected")
	}
}

/*
Build and execute an HTTP Request and frees all the resources getting all the data before
*/
func DoHttpRequestJSON(httpClient *http.Client, url string, httpMethod string, jsonToSend interface{},
	jsonToGet interface{}) (error, int) {

	req, err := http.NewRequest(httpMethod, url, ToJSONBuffer(jsonToSend))
	if err != nil {
		log.Errorf(util.LogTag("DoHttp")+"Error building request: %s", err)
		return err, -1
	}

	resp, err := httpClient.Do(req)
	if resp != nil && resp.Body != nil { // Closes the HTTP connection to the server freeing the socket files
		defer resp.Body.Close()
	}

	if err == nil { // HTTP request went well (at least at Http level)
		if jsonToGet != nil { // We want to obtain a JSON from Http body
			if resp.Body != nil { // The HTTP body HAS content
				err := json.NewDecoder(resp.Body).Decode(jsonToGet)
				if err == nil {
					return nil, resp.StatusCode
				} else {
					log.Errorf(util.LogTag("DoHttp")+"Response JSON decode error: %s", err)
					return fmt.Errorf("decoding json problems"), resp.StatusCode
				}
			} else {
				log.Errorf(util.LogTag("DoHttp") + "Empty body when expecting content")
				return fmt.Errorf("empty HTTP body"), resp.StatusCode
			}
		} else {
			return nil, resp.StatusCode
		}
	} else {
		log.Errorf(util.LogTag("DoHttp")+"HTTP error: %s", err)
		return err, -1
	}
}

/*
Encodes a golang struct into a buffer using JSON format.
*/
func ToJSONBuffer(jsonToEncode interface{}) *bytes.Buffer {
	if jsonToEncode == nil {
		return new(bytes.Buffer) // HACK!!! I should try to tell http that the body should be empty
	} else {
		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(jsonToEncode)
		return buffer
	}
}

/*
Encodes a golang struct into a an array of bytes.
*/
func ToJSONBytes(jsonToEncode interface{}) []byte {
	return ToJSONBuffer(jsonToEncode).Bytes()
}

/*
Build a valid HTTP/HTTPS URL and returns it as a string.
*/
func BuildHttpURL(https bool, ip string, port int, uri string) string {
	var protocol string
	if https {
		protocol = "https"
	} else {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s:%d%s", protocol, ip, port, uri)
}
