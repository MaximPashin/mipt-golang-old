package middleware

import (
	"log"
	"net/http"
)

type RecoverMiddleware struct {
	handler *http.Handler
	logger  *log.Logger
}

func (recMware RecoverMiddleware) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			recMware.logger.Printf("[ERROR] Panic caught: %v", rec)
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte("Internal Server Error\n"))
		}
	}()
	(*(recMware.handler)).ServeHTTP(writer, request)
}

func Recover(logger *log.Logger) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return RecoverMiddleware{handler: &handler, logger: logger}
	}
}
