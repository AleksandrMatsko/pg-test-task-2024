package api

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"pg-test-task-2024/internal/config"
)

type cmdReceivedResponse struct {
	Id string `json:"id"`
}

func cmdReceiveHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	if r.Header.Get("Content-Type") != "text/plain" {
		logger.Printf("Invalid content type: %s", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bytes, _ := io.ReadAll(r.Body)
	// TODO: check if data is a shell script

	// TODO: save cmd text in db and retrieve id
	id := uuid.New()

	f, err := os.OpenFile(config.GetCmdDir()+id.String(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		logger.Printf("Failed to create file: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	written, err := io.WriteString(f, string(bytes))
	if err != nil {
		logger.Printf("Failed to write body to file: %v", err)
		_ = os.Remove(f.Name())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if written != len(bytes) {
		logger.Printf("Failed to write body to file: wrote %d of %d bytes expected", written, r.ContentLength)
		_ = os.Remove(f.Name())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: submit cmd to executor

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(cmdReceivedResponse{Id: id.String()})
	logger.Printf("Command added: %s", f.Name())
}
