package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

/*
Validate the htpp message content extracting the JSON into a golang structure if necessary
*/
func ReceiveJSONFromHttp(w http.ResponseWriter, r *http.Request, jsonToFill interface{}) bool {
	if r.Body != nil { // Verify if HTTP message body is not empty
		err := json.NewDecoder(r.Body).Decode(jsonToFill)
		if err == nil { // HTTP message JSON content was decoded with success
			return true
		} else {
			log.Errorf("JSON decode error: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return false
		}
	} else {
		log.Errorf("HTTP empty body when expecting content")
		http.Error(w, "empty http body", http.StatusBadRequest)
		return false
	}
}

/*
Build and execute an Http Request and frees all the resources getting all the data before
*/
func DoHttpRequestJSON(httpClient *http.Client, url string, httpMethod string, jsonToSend interface{},
	jsonToGet interface{}) (error, int) {

	req, err := http.NewRequest(httpMethod, url, ToJSONBuffer(jsonToSend))
	if err != nil {
		log.Errorf("HTTP error building request %s", err)
		return err, -1
	}

	resp, err := httpClient.Do(req)
	if resp != nil && resp.Body != nil { // Closes the http connection to the server freeing the socket files
		defer resp.Body.Close()
	}

	if err == nil { // Http request went well (at least at Http level)
		if jsonToGet != nil { // We want to obtain a JSON from Http body
			if resp.Body != nil { // The Http body HAS content
				err := json.NewDecoder(resp.Body).Decode(jsonToGet)
				if err == nil {
					return nil, resp.StatusCode
				} else {
					log.Errorf("JSON decode error: %s", err)
					return fmt.Errorf("decoding json problems"), resp.StatusCode
				}
			} else {
				log.Errorf("HTTP empty body when expecting content")
				return fmt.Errorf("empty http body"), resp.StatusCode
			}
		} else {
			return nil, resp.StatusCode
		}
	} else {
		log.Errorf("HTTP error: %s", err)
		return err, -1
	}
}

/*
Encodes a golang struct into a buffer using JSON format.
*/
func ToJSONBuffer(jsonToEncode interface{}) *bytes.Buffer {
	if jsonToEncode == nil {
		return nil
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
Build a valid Http/Https URL and returns it as a string.
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
