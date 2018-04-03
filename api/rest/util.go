package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

/*
Validate the HTTP messages content and extract it to the appropriate structure
*/
func VerifyAndExtractJson(w http.ResponseWriter, r *http.Request, jsonToFill interface{}) bool {
	if r.Body != nil { // Verify if HTTP message body is not empty
		err := json.NewDecoder(r.Body).Decode(jsonToFill)
		if err == nil { // HTTP message JSON content was decoded with success
			return true
		} else {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return false
		}
	} else {
		log.Println("Empty request body")
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return false
	}
}

/*
Build an Http URL and returns it as a string.
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
