package db

import "github.com/google/uuid"

type CommandEntity struct {
	Id         uuid.UUID     `db:"id"`
	Source     string        `db:"source"`
	Status     CommandStatus `db:"status"`
	StatusDesc string        `db:"status_desc"`
	Output     string        `db:"output"`
	ExitCode   *int          `db:"exit_code"`
	Signal     *int          `db:"signal"`
}
