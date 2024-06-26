package executor

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"log"
	"os"
	"pg-test-task-2024/internal/config"
	"pg-test-task-2024/internal/db"
	"sync"
)

// Submitter should send data to specified chan
type Submitter func(uuid2 uuid.UUID)

func SubmitterProvider(submitChan chan<- uuid.UUID) Submitter {
	return func(id uuid.UUID) {
		submitChan <- id
	}
}

type Executor struct {
	// toExecChan is a chan to which the cmd receiver sends command ids
	toExecChan <-chan uuid.UUID

	worker db.TransactionWorker

	logger *log.Logger

	// runner is a function called in separate goroutine, which will
	// execute file with exec package
	runner CmdRunner

	// mtx to protect runningCommands
	mtx sync.Mutex

	// runningCommands contains a cancel function for each running command
	runningCommands map[uuid.UUID]context.CancelFunc
}

func New(toExecChan <-chan uuid.UUID, worker db.TransactionWorker, customRunner CmdRunner) *Executor {
	defaultLogger := log.Default()
	if customRunner == nil {
		customRunner = defaultRunner
	}

	return &Executor{
		toExecChan: toExecChan,
		worker:     worker,
		logger: log.New(
			defaultLogger.Writer(),
			"executor: ",
			defaultLogger.Flags()|log.Lmsgprefix),
		runner:          customRunner,
		mtx:             sync.Mutex{},
		runningCommands: make(map[uuid.UUID]context.CancelFunc),
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
			case id, ok := <-e.toExecChan:
				if !ok {
					e.logger.Printf("stopping, because chan closed")
					return
				}
				fname := config.GetCmdDir() + id.String()
				e.logger.Printf("request to exec: %s", fname)

				runnerCtx, runnerCancel := context.WithCancel(ctx)
				stop := context.AfterFunc(runnerCtx, func() {
					err := e.worker(ctx, func(tx pgx.Tx) error {
						err := db.SetCommandFailed(ctx, tx, id, "canceled")
						if err != nil {
							return err
						}
						return tx.Commit(ctx)
					})
					if err != nil {
						e.logger.Printf("failed to set command %s canceled: %s", id, err)
					} else {
						e.logger.Printf("set command %s canceled", id)
					}
				})

				e.mtx.Lock()
				e.runningCommands[id] = runnerCancel
				e.mtx.Unlock()
				go func() {
					defer os.Remove(fname)
					defer runnerCancel()

					// run the command
					e.runner(runnerCtx, id, e.worker)

					stop()

					e.mtx.Lock()
					delete(e.runningCommands, id)
					e.mtx.Unlock()
				}()
			}
		}
	}()
}

func (e *Executor) CancelCmd(id uuid.UUID) error {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	cancel, ok := e.runningCommands[id]
	if !ok {
		return ErrNotFound
	}
	e.logger.Printf("request to cancel command %s", id.String())
	cancel()
	return nil
}
