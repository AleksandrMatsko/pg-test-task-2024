package api

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"net/http"
	"pg-test-task-2024/internal/db"
)

type cmdListDto struct {
	CmdList []shortCmdDto `json:"cmd-list"`
}

type shortCmdDto struct {
	Id         uuid.UUID `json:"id"`
	Status     string    `json:"status"`
	StatusDesc string    `json:"status-desc"`
	ExitCode   *int      `json:"exit-code,omitempty"`
	Signal     *int      `json:"signal,omitempty"`
}

func toShortCmdDto(entity db.CommandEntity) shortCmdDto {
	return shortCmdDto{
		Id:         entity.Id,
		Status:     string(entity.Status),
		StatusDesc: entity.StatusDesc,
		ExitCode:   entity.ExitCode,
		Signal:     entity.Signal,
	}
}

func toCmdList(entities []db.CommandEntity) []shortCmdDto {
	list := make([]shortCmdDto, 0, len(entities))
	for _, entity := range entities {
		list = append(list, toShortCmdDto(entity))
	}
	return list
}

func getCmdListHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	encoder := json.NewEncoder(w)

	ctx := r.Context()
	var dtos []shortCmdDto
	err := doTransactional(ctx, func(tx pgx.Tx) error {
		entities, err := db.GetCommandsShortened(ctx, tx)
		if err != nil {
			return err
		}
		dtos = toCmdList(entities)
		return tx.Commit(ctx)
	})
	if err != nil {
		logger.Printf("failed to get command info: %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = encoder.Encode(errResponse{
			ShortDesc: "Internal Server Error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = encoder.Encode(cmdListDto{CmdList: dtos})
	logger.Printf("OK, send %v records", len(dtos))
}
