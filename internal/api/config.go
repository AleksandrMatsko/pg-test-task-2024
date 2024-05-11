package api

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"net/http"
)

// transactionWorker is type for function used in http handlers to work with DB.
// It should begin transaction and rollback it. If you need to commit changes
// use tx.Commit directly.
//
// Note that tx.Rollback should not affect successful commit.
type transactionWorker func(ctx context.Context, worker func(tx pgx.Tx) error) error

var doTransactional transactionWorker

func ConfigureEndpoints(starter transactionWorker) *mux.Router {
	doTransactional = starter

	r := mux.NewRouter()

	r.NotFoundHandler = notFoundHandler{}
	r.MethodNotAllowedHandler = methodNotAllowedHandler{}

	r.HandleFunc("/api/v1/ping", pingHandler)
	r.HandleFunc("/api/v1/cmd", cmdReceiveHandler).Methods(http.MethodPost)

	return r
}
