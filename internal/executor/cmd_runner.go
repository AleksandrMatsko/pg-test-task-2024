package executor

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"pg-test-task-2024/internal/db"
	"syscall"
)

type CmdRunner func(
	ctx context.Context,
	fname string,
	worker db.TransactionWorker)

func defaultRunner(ctx context.Context, fname string, worker db.TransactionWorker) {
	defaultLogger := log.Default()
	logger := log.New(
		defaultLogger.Writer(),
		fmt.Sprintf("defaultRunner %s: ", fname),
		defaultLogger.Flags()|log.Lmsgprefix)

	s, err := exec.LookPath("/bin/bash")
	if err != nil {
		logger.Printf("failed to look path: %s", err)
		return
	}
	cmd := exec.CommandContext(ctx, s, fname)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Printf("failed to get stdout pipe: %s", err)
		return
	}
	err = cmd.Start()
	if err != nil {
		logger.Printf("failed to start command: %T %s", err, err)
		return
	}

	buffered := bufio.NewReader(stdout)
	for {
		_, err := buffered.WriteTo(os.Stdout)
		if err != nil {
			logger.Printf("failed to write bytes from cmd stdout to this stdout: %s", err)
			break
		}
	}

	err = cmd.Wait()
	if err != nil {
		logger.Printf("cmd.Wait err: %s", err)
	}

	processState := cmd.ProcessState
	if processState.Exited() {
		logger.Printf("finished with exit code: %v", processState.ExitCode())
	} else {
		status := processState.Sys().(syscall.WaitStatus)
		logger.Printf("ended with signal: %v", status.Signal())
	}
}
