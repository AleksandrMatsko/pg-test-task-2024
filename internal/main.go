package internal

import (
	"context"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pg-test-task-2024/internal/api"
	"pg-test-task-2024/internal/config"
	"pg-test-task-2024/internal/db"
	"pg-test-task-2024/internal/db/migrations"
	"pg-test-task-2024/internal/executor"
	"time"
)

func Main() {
	toExecChan := make(chan uuid.UUID, 1)
	defer close(toExecChan)

	log.Println("starting server")
	log.Println("prepare directory for commands...")
	err := config.PrepareCmdDir(config.GetCmdDir())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	migrations.Apply()

	pool, err := pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// TODO: mark all running commands as error commands and try to remove files

	exe := executor.New(toExecChan, db.TransactionWorkerProvider(pool), nil)
	exe.Start(ctx)

	host := config.GetHost()
	port := config.GetPort()

	log.Printf("configuring endpoints...")
	r := api.ConfigureEndpoints(
		db.TransactionWorkerProvider(pool),
		executor.SubmitterProvider(toExecChan),
	)
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: r,
	}

	go func() {
		log.Printf("listening ...")
		err := server.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	log.Println("Shutting down...")

	ctx, cancelByTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelByTimeout()
	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("Error shutting down server gracefully: %v", err)
	}
}
