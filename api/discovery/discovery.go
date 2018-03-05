package discovery

import (
	"encoding/json"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/overlay"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
)

func ChordStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("REST WORKING")
}

func ChordLookup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fmt.Println(node.NewGuidString(params["key"]).GetBytes())
	var vnodes, _ = overlay.Overlay.Lookup(1, node.NewGuidString(params["key"]).GetBytes())
	json.NewEncoder(w).Encode(vnodes[0].Host)
}
