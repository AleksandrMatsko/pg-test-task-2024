package api

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"pg-test-task-2024/internal/db"
	"pg-test-task-2024/internal/executor"
)

var doTransactional db.TransactionWorker
var submit executor.Submitter
var cancelById func(id uuid.UUID) error

func ConfigureEndpoints(starter db.TransactionWorker, submitter executor.Submitter, cancelByIdFunc func(id uuid.UUID) error) *mux.Router {
	doTransactional = starter
	submit = submitter
	cancelById = cancelByIdFunc

	r := mux.NewRouter()

	r.NotFoundHandler = notFoundHandler{}
	r.MethodNotAllowedHandler = methodNotAllowedHandler{}

	r.HandleFunc("/api/v1/ping", pingHandler)
	r.HandleFunc("/api/v1/cmd", cmdReceiveHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/cmd/{id}/cancel", cmdCancelHandler).Methods(http.MethodPatch)

	return r
}
