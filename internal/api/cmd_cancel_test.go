package api

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"pg-test-task-2024/internal/executor"
	"testing"
)

func TestCmdCancel_WithBadUrl(t *testing.T) {
	req := httptest.NewRequest("PATCH", "/api/v1/cmd/not-uuid/cancel", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdCancelHandler)
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

func TestCmdCancel_WithCancelReturnsNil(t *testing.T) {
	cancelById = func(id uuid.UUID) error {
		return nil
	}
	id := uuid.New()

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/cmd/%s/cancel", id.String()), nil)
	req = mux.SetURLVars(req, map[string]string{
		"id": id.String(),
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdCancelHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "" {
		t.Fatalf("expected no Content-Type: got %v", contentType)
	}
}

func TestCmdCancel_WithCancelReturnsErrNotFound(t *testing.T) {
	cancelById = func(id uuid.UUID) error {
		return executor.ErrNotFound
	}
	id := uuid.New()

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/cmd/%s/cancel", id.String()), nil)
	req = mux.SetURLVars(req, map[string]string{
		"id": id.String(),
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdCancelHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
}

func TestCmdCancel_WithCancelReturnsOtherError(t *testing.T) {
	cancelById = func(id uuid.UUID) error {
		return fmt.Errorf("some other error")
	}
	id := uuid.New()

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/cmd/%s/cancel", id.String()), nil)
	req = mux.SetURLVars(req, map[string]string{
		"id": id.String(),
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cmdCancelHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("handler returned wrong content type: got %v want %v", contentType, "application/json")
	}
}
