package executor

import (
	"context"
	"github.com/google/uuid"
	"pg-test-task-2024/internal/db"
	"sync"
	"testing"
)

func TestExecutor_CallsRunner(t *testing.T) {
	execChan := make(chan uuid.UUID)
	defer close(execChan)

	var stubTransactionWorker db.TransactionWorker

	wg := sync.WaitGroup{}
	wg.Add(1)

	runCount := 0
	var gotId uuid.UUID
	stubRunner := func(ctx context.Context, id uuid.UUID, worker db.TransactionWorker) {
		runCount += 1
		gotId = id
		wg.Done()
	}

	exe := New(execChan, stubTransactionWorker, stubRunner)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exe.Start(ctx)

	expectedId := uuid.New()

	execChan <- expectedId

	wg.Wait()

	if runCount != 1 {
		t.Errorf("runner executed %d times, want %d", runCount, 1)
	}
	if gotId != expectedId {
		t.Errorf("got id %s, expected %s", gotId, expectedId)
	}

	exe.mtx.Lock()
	if len(exe.runningCommands) != 0 {
		t.Errorf("runner after execution have %d commands, expected 0", len(exe.runningCommands))
	}
	exe.mtx.Unlock()
}

func TestExecutor_CancelsRunners_WhenCanceled(t *testing.T) {
	execChan := make(chan uuid.UUID)
	defer close(execChan)
	var stubTransactionWorker db.TransactionWorker
	numRunners := 10

	wgRunnerEntered := sync.WaitGroup{}
	wgRunnerEntered.Add(numRunners)

	wgAfterCtx := sync.WaitGroup{}
	wgAfterCtx.Add(numRunners)

	ids := make([]uuid.UUID, 0, numRunners)
	for i := 0; i < numRunners; i++ {
		ids = append(ids, uuid.New())
	}

	stubRunner := func(ctx context.Context, id uuid.UUID, worker db.TransactionWorker) {
		wgRunnerEntered.Done()
		<-ctx.Done()
		wgAfterCtx.Done()
	}

	exe := New(execChan, stubTransactionWorker, stubRunner)
	ctx, cancel := context.WithCancel(context.Background())

	exe.Start(ctx)
	for i := 0; i < numRunners; i++ {
		execChan <- ids[i]
	}
	wgRunnerEntered.Wait()

	exe.mtx.Lock()
	if len(exe.runningCommands) != numRunners {
		t.Errorf("runner executing %d commands, expected %d", len(exe.runningCommands), numRunners)
	}
	exe.mtx.Unlock()

	cancel()
	wgAfterCtx.Wait()
}
