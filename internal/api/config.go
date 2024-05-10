package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

func ConfigureEndpoints() *mux.Router {
	r := mux.NewRouter()

	r.NotFoundHandler = notFoundHandler{}
	r.MethodNotAllowedHandler = methodNotAllowedHandler{}

	r.HandleFunc("/api/v1/ping", pingHandler)
	r.HandleFunc("/api/v1/cmd", cmdReceiveHandler).Methods(http.MethodPost)

	return r
}
