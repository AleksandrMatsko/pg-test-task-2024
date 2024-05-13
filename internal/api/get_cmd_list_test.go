package api

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"net/http/httptest"
	"pg-test-task-2024/internal/config"
	"pg-test-task-2024/internal/db"
	"pg-test-task-2024/internal/db/dbtest"
	"testing"
	"time"
)

func TestGetCmdList_WithDBDown(t *testing.T) {
	ctx := context.Background()
	container := dbtest.CreateTestContainer(ctx, t)

	pool, err := pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		t.Fatalf("failed to connect to testcontainer db: %s", err)
	}
	doTransactional = db.TransactionWorkerProvider(pool)

	req := httptest.NewRequest("GET", "/api/v1/cmd", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getCmdListHandler)

	timeout := time.Minute
	err = container.Stop(ctx, &timeout)
	if err != nil {
		t.Fatalf("failed to stop testcontainer: %s", err)
	}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
}

func TestGetCmdList(t *testing.T) {
	ctx := context.Background()
	dbtest.CreateTestContainer(ctx, t)

	pool, err := pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		t.Fatalf("failed to connect to testcontainer db: %s", err)
	}
	doTransactional = db.TransactionWorkerProvider(pool)

	var id uuid.UUID
	err = doTransactional(ctx, func(tx pgx.Tx) error {
		newId, err := db.InsertNewCommand(ctx, tx, correctScript)
		if err != nil {
			return err
		}
		id = newId
		return tx.Commit(ctx)
	})
	if err != nil {
		t.Fatalf("failed to insert test command into test db: %s", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/cmd", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getCmdListHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}

	decoder := json.NewDecoder(rr.Body)
	var gotDto cmdListDto
	err = decoder.Decode(&gotDto)
	if err != nil {
		t.Fatalf("failed to decode dto")
	}

	if len(gotDto.CmdList) != 1 {
		t.Fatalf("got %d commands, expected 1", len(gotDto.CmdList))
	}

	if gotDto.CmdList[0].Id != id {
		t.Fatalf("ids do not match: got %v, expected %v", gotDto.CmdList[0].Id, id)
	}
	if gotDto.CmdList[0].Status != string(db.Running) {
		t.Fatalf("status do not match: got %v, expected %v", gotDto.CmdList[0].Status, db.Running)
	}
	if gotDto.CmdList[0].StatusDesc != "" {
		t.Fatalf("status do not match: got %v, expected %v", gotDto.CmdList[0].StatusDesc, "")
	}
	if gotDto.CmdList[0].Signal != nil {
		t.Fatalf("signals do not match: got %v, expected nil", gotDto.CmdList[0].Signal)
	}
	if gotDto.CmdList[0].ExitCode != nil {
		t.Fatalf("exit code do not match: got %v, expected nil", gotDto.CmdList[0].ExitCode)
	}
}
