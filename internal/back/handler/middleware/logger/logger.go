// Package logger provides middleware that logs HTTP requests and responses,
// including the method, URI, status, response size, and the duration of the request.
package logger

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// ResponseWriter interface is redeclared here for clarity, outlining methods that
// manipulate the HTTP response.
type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

// responseData holds information about the HTTP response, including
// the status code and the size of the response in bytes.
type responseData struct {
	status int // HTTP status code of the response
	size   int // Size of the response in bytes
}

// loggingResponseWriter is a wrapper around http.ResponseWriter that captures
// the status code and response size for logging purposes.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Hijack allows the loggingResponseWriter to hijack the connection, typically used
// in WebSocket implementations. This is a necessary method to implement the http.Hijacker interface.
func (r *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("websocket: response does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}

// WithLogging returns a middleware handler that logs details about each HTTP request and its response.
// It logs the URI, method, duration, status, and response size.
//
// Parameters:
//   - h: The HTTP handler to be wrapped by the middleware.
//   - log: Logger for capturing and logging request and response details.
//
// Returns:
//   - An HTTP handler that logs the details of each request and response.
func WithLogging(h http.Handler, log *slog.Logger) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		log.Info("query: ",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
