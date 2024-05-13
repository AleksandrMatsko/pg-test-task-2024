package db

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"syscall"
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

func SetCommandFinished(ctx context.Context, tx pgx.Tx, id uuid.UUID, status syscall.WaitStatus) error {
	var err error
	if status.Exited() {
		_, err = tx.Exec(ctx, `
			UPDATE commands SET status = $1, exit_code = $2 
				WHERE id = $3
			`, Finished, status.ExitStatus(), uuid.NullUUID{UUID: id, Valid: true})
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE commands SET status = $1, signal = $2 
				WHERE id = $3
			`, Finished, status.Signal(), uuid.NullUUID{UUID: id, Valid: true})
	}
	return err
}

func SetCommandFailed(ctx context.Context, tx pgx.Tx, id uuid.UUID, description string) error {
	_, err := tx.Exec(ctx, `
			UPDATE commands SET status = $1, status_desc = $2 
				WHERE id = $3
			`, Error, description, uuid.NullUUID{UUID: id, Valid: true})
	return err
}

func AppendCommandOutput(ctx context.Context, tx pgx.Tx, id uuid.UUID, output string) error {
	_, err := tx.Exec(ctx, `
			UPDATE commands SET output = COALESCE(output, '') || $1
				WHERE id = $2
			`, output, uuid.NullUUID{UUID: id, Valid: true})
	return err
}

func GetSingleCommand(ctx context.Context, tx pgx.Tx, id uuid.UUID) (CommandEntity, error) {
	var resEntity CommandEntity
	err := tx.QueryRow(ctx, `
		SELECT * FROM commands WHERE id = $1
		`, uuid.NullUUID{UUID: id, Valid: true}).
		Scan(
			&resEntity.Id,
			&resEntity.Source,
			&resEntity.Status,
			&resEntity.StatusDesc,
			&resEntity.Output,
			&resEntity.ExitCode,
			&resEntity.Signal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CommandEntity{}, ErrEntityNotFound
		}
		return CommandEntity{}, err
	}
	return resEntity, nil
}
