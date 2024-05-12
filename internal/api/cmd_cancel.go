package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"pg-test-task-2024/internal/executor"
)

func cmdCancelHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	encoder := json.NewEncoder(w)

	s := mux.Vars(r)["id"]
	id, err := uuid.Parse(s)
	if err != nil {
		logger.Printf("invalid UUID: %s", s)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Bad Request",
			LongDesc:  fmt.Sprintf("Invalid url"),
		})
		return
	}

	err = cancelById(id)
	if err != nil {
		logger.Printf("failed to cancel command: %s", err)
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, executor.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = encoder.Encode(errResponse{
				ShortDesc: "Not Found",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_ = encoder.Encode(errResponse{
				ShortDesc: "Internal Server Error",
			})
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
