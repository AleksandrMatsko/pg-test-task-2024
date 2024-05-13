package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func MarkRunningCmdsError(ctx context.Context, tx pgx.Tx) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `
		UPDATE commands
		SET status = 'error', status_desc = 'server got down' 
		WHERE status = 'running'
		RETURNING id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
