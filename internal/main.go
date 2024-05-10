package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pg-test-task-2024/internal/api"
	"pg-test-task-2024/internal/config"
	"strings"
	"time"
)

func Main() {
	log.Println("starting server")
	log.Println("prepare directory for commands...")
	err := config.PrepareCmdDir(config.GetCmdDir())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err = pgxpool.Connect(ctx, config.GetDbConnStr())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	after, _ := strings.CutPrefix(config.GetDbConnStr(), "postgres")

	m, err := migrate.New("file://scripts/migrations", "pgx"+after)
	if err != nil {
		log.Fatalf("failed to prepare migrations: %v", err)
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	host := config.GetHost()
	port := config.GetPort()

	log.Printf("configuring endpoints...")
	r := api.ConfigureEndpoints()
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
