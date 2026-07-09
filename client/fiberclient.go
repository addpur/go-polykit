package client

import (
	"context"
	"fmt"

	"github.com/addpur/go-polykit"
	"github.com/gofiber/fiber/v2"
)

// FiberEncodeRequestFunc encodes a user-domain request into a Fiber Agent request.
type FiberEncodeRequestFunc func(context.Context, *fiber.Agent, interface{}) error

// FiberDecodeResponseFunc extracts a user-domain response from a Fiber Agent response.
type FiberDecodeResponseFunc func(context.Context, *fiber.Agent, interface{}) (interface{}, error)

// NewFiberClient creates a polykit.Endpoint that makes HTTP requests using Fiber's Agent (FastHTTP).
func NewFiberClient(
	method string,
	url string,
	enc FiberEncodeRequestFunc,
	dec FiberDecodeResponseFunc,
) polykit.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		agent := fiber.AcquireAgent()
		defer fiber.ReleaseAgent(agent)

		agent.Request().Header.SetMethod(method)
		agent.Request().SetRequestURI(url)

		if err := enc(ctx, agent, request); err != nil {
			return nil, fmt.Errorf("encode error: %w", err)
		}

		if err := agent.Parse(); err != nil {
			return nil, fmt.Errorf("parse error: %w", err)
		}

		statusCode, _, errs := agent.Bytes()
		if len(errs) > 0 {
			// Catch network errors, timeouts, etc.
			return polykit.StandardResponse{
				ResponseCode: "99",
				Message:      "Connection Timeout/Refused",
				Data:         fmt.Sprintf("internal error: %v", errs[0]),
			}, nil
		}

		if statusCode < 200 || statusCode >= 300 {
			return polykit.StandardResponse{
				ResponseCode: fmt.Sprintf("%d", statusCode),
				Message:      "Unexpected status code from upstream",
			}, nil
		}

		// Agent retains the response body and properties, we pass it to decode
		return dec(ctx, agent, request)
	}
}
