package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest/discovery"
	"github.com/strabox/caravela/node/local"
	"log"
	"net/http"
)

const API_DEBUG_ENDPOINT = "/debug"

var router *mux.Router = nil


func Initialize(apiPort int, thisNode local.LocalNode) {
	fmt.Println("[API] Initializing CARAVELA API ...")
	
	router = mux.NewRouter()
	// Endpoint used to know everything about the node (Debug Purposes Only)
	router.HandleFunc(API_DEBUG_ENDPOINT, debug).Methods("GET")
	
	// Initialize all the API endpoints
	discovery.InitializeDiscoveryAPI(router, thisNode)
	
	// Start listening hor HTTP requests
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", apiPort), router))
	
	fmt.Println("[API] API SUCCESS")
}

func debug(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Debug Endpoint")
}
