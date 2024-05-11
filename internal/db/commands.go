package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func InsertNewCommand(ctx context.Context, tx pgx.Tx, source string) (uuid.UUID, error) {
	var id uuid.NullUUID
	err := tx.QueryRow(ctx, `
		INSERT INTO commands (source, status) VALUES ($1, $2) RETURNING id
		`, source, Running).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	if !id.Valid {
		return uuid.Nil, ErrInvalidUUID
	}
	return id.UUID, nil
}
