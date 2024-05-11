package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"io"
	"net/http"
	"os"
	"pg-test-task-2024/internal/config"
	"pg-test-task-2024/internal/db"
	"regexp"
	"time"
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Unsupported Media Type",
			LongDesc:  fmt.Sprintf("Bad Content-Type, expected text/plain, got %s", contentType),
		})
		return
	}

	bytes, _ := io.ReadAll(r.Body)
	src := string(bytes)
	if !isShellScript(src) {
		logger.Printf("Not a shell script")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Bad Request",
			LongDesc:  "Not a shell script",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()
	err := doTransactional(ctx, func(tx pgx.Tx) error {
		id, err := db.InsertNewCommand(ctx, tx, src)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = encoder.Encode(errResponse{
				ShortDesc: "Internal Server Error",
			})
			return fmt.Errorf("failed to insert new command in db: %s", err)
		}

		// create file with -rwx------ permissions
		f, err := os.OpenFile(config.GetCmdDir()+id.String(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = encoder.Encode(errResponse{
				ShortDesc: "Internal Server Error",
			})
			return fmt.Errorf("failed to create file: %s", err)
		}
		defer f.Close()

		written, err := io.WriteString(f, src)
		if err != nil {
			_ = os.Remove(f.Name())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = encoder.Encode(errResponse{
				ShortDesc: "Internal Server Error",
			})
			return fmt.Errorf("failed to write body to file: %s", err)
		}

		if written != len(bytes) {
			_ = os.Remove(f.Name())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = encoder.Encode(errResponse{
				ShortDesc: "Internal Server Error",
			})
			return fmt.Errorf(
				"failed to write body to file: wrote %d of %d bytes expected",
				written, r.ContentLength)
		}

		err = tx.Commit(ctx)
		if err != nil {
			_ = os.Remove(f.Name())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = encoder.Encode(errResponse{
				ShortDesc: "Internal Server Error",
			})
			return fmt.Errorf("failed to commit changes: %s", err)
		}

		// TODO: submit cmd to executor

		w.Header().Set("Content-Type", "application/json")
		_ = encoder.Encode(cmdReceivedResponse{Id: id.String()})
		logger.Printf("Command added: %s", f.Name())
		return nil
	})
	if err != nil {
		logger.Printf("failed to save command: %s", err)
	}
}
