package discovery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/strabox/go-chord"
	"net/http"
)

var Ring *chord.Ring

func ChordStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("REST WORKING")
}

func ChordLookup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var vnodes, _ = Ring.Lookup(1, []byte(params["key"]))
	json.NewEncoder(w).Encode(vnodes)
}
