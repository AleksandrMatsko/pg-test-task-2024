package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"pg-test-task-2024/internal/config"
	"pg-test-task-2024/internal/db"
	"pg-test-task-2024/internal/db/dbtest"
	"pg-test-task-2024/internal/executor"
	"strings"
	"testing"
	"time"
)

func TestCmdReceiveHandler_WithWrongContentType(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/cmd", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdReceiveHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnsupportedMediaType {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusUnsupportedMediaType)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
}

var notShellScripts = []string{
	`#/bin/bash`,
	`!/bin/bash`,
	`#!/bins/sh`,
}

func testCmdReceiveHandler_WithNotShellScript(t *testing.T, script string) {
	req := httptest.NewRequest("POST", "/api/v1/cmd", strings.NewReader(script))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdReceiveHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestCmdReceiveHandler_WithNotShellScripts(t *testing.T) {
	for i, script := range notShellScripts {
		t.Run(fmt.Sprintf("bad script %v", i),
			func(t *testing.T) {
				testCmdReceiveHandler_WithNotShellScript(t, script)
			})
	}
}

var correctScript = `#!/bin/bash
echo 'Hello world!'
`

func TestCmdReceiveHandler_WithDBDown(t *testing.T) {
	ctx := context.Background()
	container := dbtest.CreateTestContainer(ctx, t)
	pool, err := pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		t.Fatalf("failed to connect to testcontainer db: %s", err)
	}
	doTransactional = db.TransactionWorkerProvider(pool)
	execChan := make(chan uuid.UUID, 1)
	submit = executor.SubmitterProvider(execChan)
	t.Cleanup(func() {
		doTransactional = nil
		submit = nil
	})

	req := httptest.NewRequest("POST", "/api/v1/cmd", strings.NewReader(correctScript))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdReceiveHandler)

	timeout := time.Minute * 2
	err = container.Stop(ctx, &timeout)
	if err != nil {
		t.Fatalf("failed to stop testcontainer: %s", err)
	}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}

	close(execChan)
	gotId := <-execChan
	if gotId != uuid.Nil {
		t.Fatalf("no data should be send to executor: got %v", gotId)
	}
}

func TestCmdReceiveHandler_WithNoCmdDir(t *testing.T) {
	t.Setenv("EXECUTOR_CMD_DIR", "/tmp/not_exist_commands/")

	ctx := context.Background()
	dbtest.CreateTestContainer(ctx, t)
	pool, err := pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		t.Fatalf("failed to connect to testcontainer db: %s", err)
	}
	doTransactional = db.TransactionWorkerProvider(pool)
	execChan := make(chan uuid.UUID, 1)
	submit = executor.SubmitterProvider(execChan)
	t.Cleanup(func() {
		doTransactional = nil
		submit = nil
	})

	req := httptest.NewRequest("POST", "/api/v1/cmd", strings.NewReader(correctScript))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdReceiveHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}

	rows, err := pool.Query(ctx, `SELECT * FROM commands`)
	if err != nil {
		t.Fatalf("failed to get commands from testcontainer db: %s", err)
	}
	if rows.Next() {
		t.Fatalf("expected no rows in db")
	}

	close(execChan)
	gotId := <-execChan
	if gotId != uuid.Nil {
		t.Fatalf("no data should be send to executor: got %v", gotId)
	}
}

// TestCmdReceiveHandler_WithShellScript tests full user scenario
//   - request received
//   - data inserted to db
//   - file with script created
func TestCmdReceiveHandler_WithShellScript(t *testing.T) {
	err := config.PrepareCmdDir(config.GetCmdDir())
	if err != nil {
		t.Fatalf("failed to prepare cmd dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(config.GetCmdDir())
	})

	ctx := context.Background()
	dbtest.CreateTestContainer(ctx, t)
	pool, err := pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		t.Fatalf("failed to connect to testcontainer db: %s", err)
	}
	doTransactional = db.TransactionWorkerProvider(pool)
	execChan := make(chan uuid.UUID, 1)
	submit = executor.SubmitterProvider(execChan)
	t.Cleanup(func() {
		doTransactional = nil
		submit = nil
	})

	req := httptest.NewRequest("POST", "/api/v1/cmd", strings.NewReader(correctScript))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdReceiveHandler)

	handler.ServeHTTP(rr, req)

	// check that status code is correct
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// check that response is a JSON with correct Content-Type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
	decoder := json.NewDecoder(rr.Body)
	var rsp cmdReceivedResponse
	err = decoder.Decode(&rsp)
	if err != nil {
		t.Fatalf("unexpected error decoding response: %v", err)
	}

	// check that id is a uuid
	parsedId, err := uuid.Parse(rsp.Id)
	if err != nil {
		t.Fatalf("unexpected error parsing id: %v", err)
	}

	// check that file is created
	expectedFileName := config.GetCmdDir() + rsp.Id
	f, err := os.Open(expectedFileName)
	if err != nil {
		t.Fatalf("unexpected error opening command file: %v", err)
	}
	defer f.Close()
	t.Cleanup(func() {
		_ = os.Remove(expectedFileName)
	})

	bytes, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("unexpected error reading command file: %v", err)
	}
	if string(bytes) != correctScript {
		t.Fatalf("command returned wrong content:\ngot\n%v\n--------\nwant\n%v\n", string(bytes), correctScript)
	}

	var resEntity db.CommandEntity
	err = doTransactional(ctx, func(tx pgx.Tx) error {
		entity, err := db.GetSingleCommand(ctx, tx, parsedId)
		if err != nil {
			return err
		}
		resEntity = entity
		return nil
	})
	if err != nil {
		t.Fatalf("failed to get inserted entity: %s", err)
	}

	checkCommandsEntities(t, resEntity, db.CommandEntity{
		Id:     parsedId,
		Source: correctScript,
		Status: db.Running,
	})

	close(execChan)
	gotId := <-execChan
	if gotId != parsedId {
		t.Fatalf("expected sending id to executor's chan: got %v expected %v", gotId, parsedId)
	}
}

func checkCommandsEntities(t *testing.T, got, expected db.CommandEntity) {
	if got.Id != expected.Id {
		t.Fatalf("ids do not match: got %v, expected %v", got.Id, expected.Id)
	}
	if got.Source != expected.Source {
		t.Fatalf("sources do not match:\ngot:\n%v\n----------\nexpected:\n%v\n", got.Source, expected.Source)
	}
	if got.Status != expected.Status {
		t.Fatalf("statuses do not match: got %v, expected %v", got.Status, expected.Status)
	}
	if got.StatusDesc != expected.StatusDesc {
		t.Fatalf("status descriptions do not match: got %v, expected %v", got.StatusDesc, expected.StatusDesc)
	}
	if got.Output != expected.Output {
		t.Fatalf("outputs do not match:\ngot:\n%v\n----------\nexpected:\n%v\n", got.Output, expected.Output)
	}
	if got.ExitCode != expected.ExitCode {
		t.Fatalf("exit codes do not match: got %v, expected %v", got.ExitCode, expected.ExitCode)
	}
	if got.Signal != expected.Signal {
		t.Fatalf("signals do not match: got %v, expected %v", got.Signal, expected.Signal)
	}
}
