package api

import (
	"encoding/json"
	"net/http"
)

type errResponse struct {
	ShortDesc string `json:"short-desc"`
	LongDesc  string `json:"long-desc,omitempty"`
}

type notFoundHandler struct{}

func (notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	logger.Printf("Not Found")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(errResponse{ShortDesc: "Not Found"})
}

type methodNotAllowedHandler struct{}

func (methodNotAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	logger.Printf("Method Not Allowed")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(errResponse{ShortDesc: "Method Not Allowed"})
}
