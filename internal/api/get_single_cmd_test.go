package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

func TestGetSingleCmd_WithBadUrl(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/cmd/not-uuid", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getSingleCmdHandler)
	req = mux.SetURLVars(req, map[string]string{
		"id": "not-uuid",
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
}

func TestGetSingleCmd_WithDBDown(t *testing.T) {
	ctx := context.Background()
	container := dbtest.CreateTestContainer(ctx, t)

	pool, err := pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		t.Fatalf("failed to connect to testcontainer db: %s", err)
	}
	doTransactional = db.TransactionWorkerProvider(pool)

	id := uuid.New()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/cmd/%s", id), nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getSingleCmdHandler)
	req = mux.SetURLVars(req, map[string]string{
		"id": id.String(),
	})

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

func TestGetSingleCmd_WithNoCmdInDB(t *testing.T) {
	ctx := context.Background()
	dbtest.CreateTestContainer(ctx, t)

	pool, err := pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		t.Fatalf("failed to connect to testcontainer db: %s", err)
	}
	doTransactional = db.TransactionWorkerProvider(pool)

	id := uuid.New()
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/cmd/%s", id), nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getSingleCmdHandler)
	req = mux.SetURLVars(req, map[string]string{
		"id": id.String(),
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
}

func TestGetSingleCmd_WithCmdInDB(t *testing.T) {
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

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/cmd/%s", id), nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getSingleCmdHandler)
	req = mux.SetURLVars(req, map[string]string{
		"id": id.String(),
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}

	decoder := json.NewDecoder(rr.Body)
	var gotDto singleCmdDto
	err = decoder.Decode(&gotDto)
	if err != nil {
		t.Fatalf("failed to decode dto")
	}

	if gotDto.Id != id {
		t.Fatalf("ids do not match: got %v, expected %v", gotDto.Id, id)
	}
	if gotDto.Source != correctScript {
		t.Fatalf("sources do not match: got %v, expected %v", gotDto.Source, correctScript)
	}
	if gotDto.Status != string(db.Running) {
		t.Fatalf("status do not match: got %v, expected %v", gotDto.Status, db.Running)
	}
	if gotDto.StatusDesc != "" {
		t.Fatalf("status do not match: got %v, expected %v", gotDto.StatusDesc, "")
	}
	if gotDto.Output != "" {
		t.Fatalf("outputs do not match: got %v, expected %v", gotDto.Output, "")
	}
	if gotDto.Signal != nil {
		t.Fatalf("signals do not match: got %v, expected nil", gotDto.Signal)
	}
	if gotDto.ExitCode != nil {
		t.Fatalf("exit code do not match: got %v, expected nil", gotDto.ExitCode)
	}
}
