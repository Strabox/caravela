package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/discovery"
	"log"
	"net/http"
	"runtime"
	"strconv"
)

const CARAVELA_PORT = 8000

func main() {
	fmt.Println("Starting CARAVELA")
	fmt.Println("#cores: " + strconv.Itoa(runtime.NumCPU()))

	router := mux.NewRouter()
	router.HandleFunc("/people", discovery.SearchResources).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(CARAVELA_PORT), router))

	fmt.Println("Stopping CARAVELA")
}
