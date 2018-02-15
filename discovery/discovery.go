package discovery

import (
	"encoding/json"
	"net/http"
)

func SearchResources(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("a")
}
