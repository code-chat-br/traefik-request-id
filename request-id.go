package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type Config struct {
	HeaderName string `json:"headerName,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		HeaderName: "X-Request-ID",
	}
}

type RequestID struct {
	next       http.Handler
	name       string
	headerName string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	headerName := config.HeaderName

	if headerName == "" {
		headerName = "X-Request-ID"
	}

	return &RequestID{
		next:       next,
		name:       name,
		headerName: headerName,
	}, nil
}

func (r *RequestID) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestID := req.Header.Get(r.headerName)

	if requestID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			http.Error(rw, "failed to generate request id", http.StatusInternalServerError)
			return
		}

		requestID = id.String()
		req.Header.Set(r.headerName, requestID)
	}

	rw.Header().Set(r.headerName, requestID)

	r.next.ServeHTTP(rw, req)
}
