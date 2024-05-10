package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"pg-test-task-2024/internal/config"
	"regexp"
)

type cmdReceivedResponse struct {
	Id string `json:"id"`
}

var shellPattern = regexp.MustCompile(`(^#!/bin/bash|^#!/bin/sh)`)

// isShellScript checks if file starts with any of sequences
//   - #!/bin/bash
//   - #!/bin/sh
//
// Check is performed with regular expressions.
// I'm sure there is more suitable way to do it.
func isShellScript(content string) bool {
	return shellPattern.MatchString(content)
}

func cmdReceiveHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	encoder := json.NewEncoder(w)

	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		logger.Printf("Invalid content type: %s, expected text/plain", contentType)
		w.WriteHeader(http.StatusBadRequest)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Bad Request",
			LongDesc:  fmt.Sprintf("Bad Content-Type, expected text/plain, got %s", contentType),
		})
		return
	}

	bytes, _ := io.ReadAll(r.Body)
	str := string(bytes)
	if !isShellScript(str) {
		logger.Printf("Not a shell script")
		w.WriteHeader(http.StatusBadRequest)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Bad Request",
			LongDesc:  "Not a shell script",
		})
		return
	}

	// TODO: save cmd text in db and retrieve id
	id := uuid.New()

	// create file with -rwx------ permissions
	f, err := os.OpenFile(config.GetCmdDir()+id.String(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		logger.Printf("Failed to create file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Internal Server Error",
		})
		return
	}
	defer f.Close()

	written, err := io.WriteString(f, str)
	if err != nil {
		logger.Printf("Failed to write body to file: %v", err)
		_ = os.Remove(f.Name())
		w.WriteHeader(http.StatusInternalServerError)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Internal Server Error",
		})
		return
	}

	if written != len(bytes) {
		logger.Printf("Failed to write body to file: wrote %d of %d bytes expected", written, r.ContentLength)
		_ = os.Remove(f.Name())
		w.WriteHeader(http.StatusInternalServerError)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Internal Server Error",
		})
		return
	}

	// TODO: submit cmd to executor

	w.Header().Set("Content-Type", "application/json")
	_ = encoder.Encode(cmdReceivedResponse{Id: id.String()})
	logger.Printf("Command added: %s", f.Name())
}
