package db

import "github.com/google/uuid"

type CommandEntity struct {
	Id         uuid.UUID
	Source     string
	Status     CommandStatus
	StatusDesc string
	Output     string
	ExitCode   *int
	Signal     *int
}
