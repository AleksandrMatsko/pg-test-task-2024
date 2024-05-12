package api

import (
	"github.com/gorilla/mux"
	"net/http"
	"pg-test-task-2024/internal/db"
	"pg-test-task-2024/internal/executor"
)

var doTransactional db.TransactionWorker
var submit executor.Submitter

func ConfigureEndpoints(starter db.TransactionWorker, submitter executor.Submitter) *mux.Router {
	doTransactional = starter
	submit = submitter

	r := mux.NewRouter()

	r.NotFoundHandler = notFoundHandler{}
	r.MethodNotAllowedHandler = methodNotAllowedHandler{}

	r.HandleFunc("/api/v1/ping", pingHandler)
	r.HandleFunc("/api/v1/cmd", cmdReceiveHandler).Methods(http.MethodPost)

	return r
}
