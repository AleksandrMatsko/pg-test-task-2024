package api

import (
	"fmt"
	"log"
	"net/http"
)

func getLogger(r *http.Request) *log.Logger {
	defaultLogger := log.Default()
	return log.New(
		defaultLogger.Writer(),
		fmt.Sprintf("%s %s: ", r.Method, r.URL),
		defaultLogger.Flags()|log.Lmsgprefix)
}
