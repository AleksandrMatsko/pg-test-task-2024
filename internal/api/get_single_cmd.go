package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"net/http"
	"pg-test-task-2024/internal/db"
)

type singleCmdDto struct {
	Id         uuid.UUID `json:"id"`
	Source     string    `json:"source"`
	Status     string    `json:"status"`
	StatusDesc string    `json:"status-desc"`
	Output     string    `json:"output"`
	ExitCode   *int      `json:"exit-code,omitempty"`
	Signal     *int      `json:"signal,omitempty"`
}

func toSingleCmdDto(entity db.CommandEntity) singleCmdDto {
	return singleCmdDto{
		Id:         entity.Id,
		Source:     entity.Source,
		Status:     string(entity.Status),
		StatusDesc: entity.StatusDesc,
		Output:     entity.Output,
		ExitCode:   entity.ExitCode,
		Signal:     entity.Signal,
	}
}

func getSingleCmdHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	encoder := json.NewEncoder(w)

	s := mux.Vars(r)["id"]
	id, err := uuid.Parse(s)
	if err != nil {
		logger.Printf("%s is invalid UUID: %s", s, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Bad Request",
			LongDesc:  fmt.Sprintf("Invalid url"),
		})
		return
	}

	ctx := r.Context()
	var rsp singleCmdDto
	err = doTransactional(ctx, func(tx pgx.Tx) error {
		entity, err := db.GetSingleCommand(ctx, tx, id)
		if err != nil {
			return err
		}
		rsp = toSingleCmdDto(entity)
		return nil
	})
	if err != nil {
		logger.Printf("failed to get command info: %T %s", err, err)
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, db.ErrEntityNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = encoder.Encode(errResponse{
				ShortDesc: "Not Found",
				LongDesc:  "Entity with such id not found",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Internal Server Error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = encoder.Encode(rsp)
	logger.Printf("OK")
}
