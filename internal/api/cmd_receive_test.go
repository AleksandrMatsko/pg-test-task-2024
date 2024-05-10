package api

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"pg-test-task-2024/internal/config"
	"strings"
	"testing"
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
	for _, script := range notShellScripts {
		testCmdReceiveHandler_WithNotShellScript(t, script)
	}
}

var correctScript = `#!/bin/bash
echo 'Hello world!'
`

func TestCmdReceiveHandler_WithNoCmdDir(t *testing.T) {
	t.Setenv("EXECUTOR_CMD_DIR", "/tmp/not_exist_commands/")

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
}

func TestCmdReceiveHandler_WithShellScript(t *testing.T) {
	err := config.PrepareCmdDir(config.GetCmdDir())
	if err != nil {
		t.Fatalf("failed to prepare cmd dir: %v", err)
	}

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
	_, err = uuid.Parse(rsp.Id)
	if err != nil {
		t.Fatalf("unexpected error parsing id: %v", err)
	}

	// check that file is created
	f, err := os.Open(config.GetCmdDir() + rsp.Id)
	if err != nil {
		t.Fatalf("unexpected error opening command file: %v", err)
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("unexpected error reading command file: %v", err)
	}
	if string(bytes) != correctScript {
		t.Fatalf("command returned wrong content:\ngot\n%v\n--------\nwant\n%v\n", string(bytes), correctScript)
	}

	_ = os.Remove(config.GetCmdDir() + rsp.Id)
	_ = os.Remove(config.GetCmdDir())
}
