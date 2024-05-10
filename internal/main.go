package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pg-test-task-2024/internal/api"
	"pg-test-task-2024/internal/config"
	"time"
)

func Main() {
	log.Println("starting server")
	log.Println("prepare directory for commands...")
	err := config.PrepareCmdDir(config.GetCmdDir())
	if err != nil {
		log.Fatal(err)
	}

	host := config.GetHost()
	port := config.GetPort()

	log.Printf("configuring endpoints...")
	r := api.ConfigureEndpoints()
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: r,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
