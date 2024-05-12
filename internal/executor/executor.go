package executor

import (
	"context"
	"github.com/google/uuid"
	"log"
	"os"
	"pg-test-task-2024/internal/db"
	"strings"
	"sync"
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

func New(toExecChan <-chan string, worker db.TransactionWorker, customRunner CmdRunner) *Executor {
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
			case fname, ok := <-e.toExecChan:
				if !ok {
					e.logger.Printf("stopping, because chan closed")
				}
				e.logger.Printf("request to exec: %s", fname)

				strs := strings.SplitAfter(fname, "/")
				id, err := uuid.Parse(strs[len(strs)-1])
				if err != nil {
					e.logger.Printf("bad file name: %s", fname)
					continue
				}

				runnerCtx, runnerCancel := context.WithCancel(ctx)

				e.mtx.Lock()
				e.runningCommands[id] = runnerCancel
				go func() {
					defer os.Remove(fname)
					defer runnerCancel()

					// run the command
					e.runner(runnerCtx, fname, e.worker)

					e.mtx.Lock()
					delete(e.runningCommands, id)
					e.mtx.Unlock()
				}()
				e.mtx.Unlock()
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
	cancel()
	return nil
}
