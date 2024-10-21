package tfetest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LoggedRequest contains information about an HTTP request and its response
type LoggedRequest struct {
	Method      string
	Path        string
	Headers     http.Header
	Body        []byte
	StatusCode  int
	Duration    time.Duration
	RequestTime time.Time
}

// RequestLogger keeps track of all requests made to the server
type RequestLogger struct {
	Requests []LoggedRequest
}

// NewRequestLogger creates a new RequestLogger
func NewRequestLogger() *RequestLogger {
	return &RequestLogger{
		Requests: make([]LoggedRequest, 0),
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default to 200 if WriteHeader is never called
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Middleware returns a middleware function that logs requests
func (rl *RequestLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture request body
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			// Restore the body for the next handler
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Wrap the response writer
		rw := newResponseWriter(w)

		// Record start time
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Create the logged request
		logged := LoggedRequest{
			Method:      r.Method,
			Path:        r.URL.Path,
			Headers:     r.Header.Clone(),
			Body:        bodyBytes,
			StatusCode:  rw.statusCode,
			Duration:    time.Since(start),
			RequestTime: start,
		}

		// Store the request
		rl.Requests = append(rl.Requests, logged)
	})
}

// LastRequest returns the most recent request, or nil if no requests have
// been made
func (rl *RequestLogger) LastRequest() *LoggedRequest {
	if len(rl.Requests) == 0 {
		return nil
	}
	return &rl.Requests[len(rl.Requests)-1]
}

// Reset clears all logged requests
func (rl *RequestLogger) Reset() {
	rl.Requests = rl.Requests[:0]
}

// RequestsForPath returns all requests made to a specific path
func (rl *RequestLogger) RequestsForPath(path string) []LoggedRequest {
	var requests []LoggedRequest
	for _, req := range rl.Requests {
		if req.Path == path {
			requests = append(requests, req)
		}
	}
	return requests
}

// String returns a formatted string representation of a LoggedRequest
func (lr *LoggedRequest) String() string {
	return fmt.Sprintf("[%s] %s %s -> %d (%s)",
		lr.RequestTime.Format(time.RFC3339),
		lr.Method,
		lr.Path,
		lr.StatusCode,
		lr.Duration)
}
