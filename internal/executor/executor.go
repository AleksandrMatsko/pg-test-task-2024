package executor

import (
	"context"
	"log"
	"pg-test-task-2024/internal/db"
)

// Submitter should send data to specified chan
type Submitter func(string)

func SubmitterProvider(submitChan chan<- string) Submitter {
	return func(fname string) {
		submitChan <- fname
	}
}

type Executor struct {
	// toExecChan is a chan to which the cmd receiver sends command ids
	toExecChan <-chan string

	doTransactional db.TransactionWorker

	logger *log.Logger

	// TODO: cmd runner
}

func New(toExecChan <-chan string, worker db.TransactionWorker) *Executor {
	defaultLogger := log.Default()

	return &Executor{
		toExecChan:      toExecChan,
		doTransactional: worker,
		logger: log.New(
			defaultLogger.Writer(),
			"executor: ",
			defaultLogger.Flags()|log.Lmsgprefix),
	}
}

// Start starts separate goroutine
func (e *Executor) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				e.logger.Printf("stopping, because context done: %s", ctx.Err())
				return
			case fname, ok := <-e.toExecChan:
				if !ok {
					e.logger.Printf("stopping, because chan closed")
				}
				e.logger.Printf("request to exec: %s", fname)
				// TODO: exec file in separate goroutine
			}
		}
	}()
}
