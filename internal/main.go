package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pg-test-task-2024/internal/api"
	"time"
)

func Main() {
	host := "0.0.0.0"
	port := 8081

	log.Printf("configuring endpoints...")
	r := api.ConfigureEndpoints()
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: r,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Printf("listening on http://%s:%d ...", host, port)
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
	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("Error shutting down server gracefully: %v", err)
	}
}
