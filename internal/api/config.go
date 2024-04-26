package api

import (
	"github.com/gorilla/mux"
)

func ConfigureEndpoints() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/ping", pingHandler)

	return r
}
