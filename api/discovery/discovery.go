package discovery

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/node"
	"github.com/strabox/caravela/node/guid"
	"net/http"
)

var thisNode *node.Node = nil

func InitializeDiscoveryAPI(selfNode *node.Node) {
	thisNode = selfNode
}

func ChordStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("REST WORKING")
}

func ChordLookup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	res := thisNode.Overlay().Lookup(*guid.NewGuidString(params["key"]))
	json.NewEncoder(w).Encode(res[0].IP())
}
