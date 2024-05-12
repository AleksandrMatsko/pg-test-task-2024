package executor

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"io"
	"log"
	"os/exec"
	"pg-test-task-2024/internal/config"
	"pg-test-task-2024/internal/db"
	"syscall"
)

func setCmdFailed(ctx context.Context, worker db.TransactionWorker, id uuid.UUID, description string) {
	_ = worker(ctx, func(tx pgx.Tx) error {
		err := db.SetCommandFailed(ctx, tx, id, description)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	})
}

type CmdRunner func(
	ctx context.Context,
	id uuid.UUID,
	worker db.TransactionWorker)

func defaultRunner(ctx context.Context, id uuid.UUID, worker db.TransactionWorker) {
	defaultLogger := log.Default()
	fname := config.GetCmdDir() + id.String()
	logger := log.New(
		defaultLogger.Writer(),
		fmt.Sprintf("defaultRunner %s: ", fname),
		defaultLogger.Flags()|log.Lmsgprefix)

	s, err := exec.LookPath("/bin/sh")
	if err != nil {
		logger.Printf("failed to look path of /bin/bash: %s", err)
		setCmdFailed(ctx, worker, id, "/bin/bash not found")
		return
	}
	cmd := exec.CommandContext(ctx, s, fname)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Printf("failed to get stdout pipe: %s", err)
		setCmdFailed(ctx, worker, id, "failed to connect to script stdout")
		return
	}
	err = cmd.Start()
	if err != nil {
		logger.Printf("failed to start command: %T %s", err, err)
		setCmdFailed(ctx, worker, id, "failed to start script")
		return
	}

	buffer := make([]byte, 1024)
	for {
		n, err := stdout.Read(buffer)
		if err != nil {
			if err != io.EOF {
				logger.Printf("failed to read stdout: %s", err)
				setCmdFailed(ctx, worker, id, "failed to read stdout")
				return
			}
		}
		if n != 0 {
			str := string(buffer[:n])
			err = worker(ctx, func(tx pgx.Tx) error {
				err := db.AppendCommandOutput(ctx, tx, id, str)
				if err != nil {
					return err
				}
				return tx.Commit(ctx)
			})
			if err != nil {
				logger.Printf("failed to append command output: %s", err)
				setCmdFailed(ctx, worker, id, "failed to append command output")
				return
			}
			logger.Printf("append %v bytes to command output", n)
		}
		if n == 0 || err == io.EOF {
			break
		}
	}

	err = cmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			logger.Printf("cmd.Wait err: %s", err)
			if errors.Is(err, context.Canceled) {
				setCmdFailed(ctx, worker, id, "cancelled")
			} else {
				setCmdFailed(ctx, worker, id, "internal error")
			}
			return
		}
	}

	processState := cmd.ProcessState
	status := processState.Sys().(syscall.WaitStatus)
	if status.Exited() {
		logger.Printf("finished with exit status: %v", status.ExitStatus())
	} else {
		logger.Printf("ended with signal: %v", status.Signal())
	}
	err = worker(ctx, func(tx pgx.Tx) error {
		err := db.SetCommandFinished(ctx, tx, id, status)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	})
	if err != nil {
		logger.Printf("failed to set command finished: %s", err)
		setCmdFailed(ctx, worker, id, "internal error")
	}
}
