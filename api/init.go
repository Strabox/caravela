package api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/discovery"
	"log"
	"net/http"
	"strconv"
)

func Initialize(apiPort int) {
	fmt.Println("[API] Initializing CARAVELA REST API ...")
	router := mux.NewRouter()
	router.HandleFunc("/debug/status", discovery.ChordStatus).Methods("GET")
	router.HandleFunc("/lookup/{key}", discovery.ChordLookup).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(apiPort), router))
	fmt.Println("[API] API Initialized Success")
}
