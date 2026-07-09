package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/addpur/go-polykit"
)

// HTTPEncodeRequestFunc encodes a user-domain request into an HTTP request.
type HTTPEncodeRequestFunc func(context.Context, *http.Request, interface{}) error

// HTTPDecodeResponseFunc extracts a user-domain response from an HTTP response.
type HTTPDecodeResponseFunc func(context.Context, *http.Response) (interface{}, error)

// NewHTTPClient creates a polykit.Endpoint that makes HTTP requests using the standard net/http client.
func NewHTTPClient(
	method string,
	url string,
	enc HTTPEncodeRequestFunc,
	dec HTTPDecodeResponseFunc,
	client *http.Client,
) polykit.Endpoint {
	if client == nil {
		client = http.DefaultClient
	}

	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request error: %w", err)
		}

		if err := enc(ctx, req, request); err != nil {
			return nil, fmt.Errorf("encode error: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("execute request error: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		return dec(ctx, resp)
	}
}
