package rest

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/util"
	"net/http"
)

/*
HTTP Generic handler for all of HTTP endpoints
*/
type AppHandler func(http.ResponseWriter, *http.Request) (interface{}, error)

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if responseJSON, err := fn(w, r); err != nil { // Handler returned an error processing the HTTP request
		log.Errorf(util.LogTag("[Handler]")+"%s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else { // All fine processing the HTTP request
		w.WriteHeader(http.StatusOK)
		if responseJSON != nil {
			w.Write(ToJSONBytes(responseJSON))
		}
	}
}
