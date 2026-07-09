package transport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/addpur/go-polykit"
)

// HTTPDecodeRequestFunc extracts a user-domain request object from an HTTP request.
type HTTPDecodeRequestFunc func(*http.Request) (interface{}, error)

// HTTPEncodeResponseFunc encodes a user-domain response object into an HTTP response.
type HTTPEncodeResponseFunc func(http.ResponseWriter, *http.Request, interface{}) error

// HTTPEncodeErrorFunc encodes a user-domain error into an HTTP response.
type HTTPEncodeErrorFunc func(http.ResponseWriter, error)

// NewHTTPServer constructs a new http.Handler for the given endpoint.
// It is fully compatible with Gorilla Mux or standard net/http mux.
func NewHTTPServer(
	e polykit.Endpoint,
	dec HTTPDecodeRequestFunc,
	enc HTTPEncodeResponseFunc,
	errEnc HTTPEncodeErrorFunc,
) http.HandlerFunc {
	if errEnc == nil {
		errEnc = DefaultHTTPErrorEncoder
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := dec(r)
		if err != nil {
			errEnc(w, err)
			return
		}

		ctx := r.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 {
				switch strings.ToLower(parts[0]) {
				case "bearer":
					ctx = context.WithValue(ctx, "auth_token", parts[1])
				case "basic":
					decoded, err := base64.StdEncoding.DecodeString(parts[1])
					if err == nil {
						ctx = context.WithValue(ctx, "auth_basic", string(decoded))
					}
				}
			}
		}

		res, err := e(ctx, req)
		if err != nil {
			errEnc(w, err)
			return
		}

		if err := enc(w, r, res); err != nil {
			errEnc(w, err)
			return
		}
	}
}

func DefaultHTTPErrorEncoder(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(polykit.StandardResponse{
		ResponseCode: "500",
		Message:      err.Error(),
	})
}
