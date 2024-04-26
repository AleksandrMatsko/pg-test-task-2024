package api

import "net/http"

func pingHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	logger.Println("pong")
	_, _ = w.Write([]byte("pong"))
}
