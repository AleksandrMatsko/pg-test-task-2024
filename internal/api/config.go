package api

import (
	"github.com/gorilla/mux"
	"net/http"
	"pg-test-task-2024/internal/db"
)

var doTransactional db.TransactionWorker

func ConfigureEndpoints(starter db.TransactionWorker) *mux.Router {
	doTransactional = starter

	r := mux.NewRouter()

	r.NotFoundHandler = notFoundHandler{}
	r.MethodNotAllowedHandler = methodNotAllowedHandler{}

	r.HandleFunc("/api/v1/ping", pingHandler)
	r.HandleFunc("/api/v1/cmd", cmdReceiveHandler).Methods(http.MethodPost)

	return r
}
