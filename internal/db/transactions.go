package db

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// TransactionWorker is type for function used in http handlers to work with DB.
// It should begin transaction and rollback it. If you need to commit changes
// use tx.Commit directly.
//
// Note that tx.Rollback should not affect successful commit.
type TransactionWorker func(ctx context.Context, worker func(tx pgx.Tx) error) error

func TransactionWorkerProvider(pool *pgxpool.Pool) TransactionWorker {
	return func(ctx context.Context, worker func(tx pgx.Tx) error) error {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		err = worker(tx)
		return err
	}
}
