package db

type CommandStatus string

const (
	Running  CommandStatus = "running"
	Error    CommandStatus = "error"
	Finished CommandStatus = "finished"
)
