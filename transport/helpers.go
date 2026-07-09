package transport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/addpur/go-polykit"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/metadata"
)

// HTTPToContext returns an HTTPRequestFunc that extracts the Authorization header
// (Bearer token or Basic credentials) and injects them into the context.
// Bearer tokens → context key "auth_token"
// Basic credentials → context key "auth_basic" (decoded "username:password")
func HTTPToContext() HTTPRequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return ctx
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			return ctx
		}
		switch strings.ToLower(parts[0]) {
		case "bearer":
			ctx = context.WithValue(ctx, "auth_token", parts[1])
		case "basic":
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err == nil {
				ctx = context.WithValue(ctx, "auth_basic", string(decoded))
			}
		}
		return ctx
	}
}

// PopulateRequestID extracts X-Request-ID from the HTTP header and injects it into context.
func PopulateRequestID() HTTPRequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		reqID := r.Header.Get("X-Request-ID")
		if reqID != "" {
			ctx = context.WithValue(ctx, "request_id", reqID)
		}
		return ctx
	}
}

// SetResponseHeader returns an HTTPResponseFunc that sets a response header.
func SetResponseHeader(key, val string) HTTPResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter) context.Context {
		w.Header().Set(key, val)
		return ctx
	}
}

// FiberToContext returns a FiberRequestFunc that extracts the Authorization header
// (Bearer token or Basic credentials) and injects them into the context.
func FiberToContext() FiberRequestFunc {
	return func(ctx context.Context, c *fiber.Ctx) context.Context {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return ctx
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			return ctx
		}
		switch strings.ToLower(parts[0]) {
		case "bearer":
			ctx = context.WithValue(ctx, "auth_token", parts[1])
		case "basic":
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err == nil {
				ctx = context.WithValue(ctx, "auth_basic", string(decoded))
			}
		}
		return ctx
	}
}

// SetFiberResponseHeader returns a FiberResponseFunc that sets a response header.
func SetFiberResponseHeader(key, val string) FiberResponseFunc {
	return func(ctx context.Context, c *fiber.Ctx) context.Context {
		c.Set(key, val)
		return ctx
	}
}

// GRPCToContext returns a GRPCRequestFunc that extracts the authorization metadata
// (Bearer token or Basic credentials) and injects them into the context.
func GRPCToContext() GRPCRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		authTokens := md.Get("authorization")
		if len(authTokens) == 0 {
			return ctx
		}
		authHeader := authTokens[0]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			return ctx
		}
		switch strings.ToLower(parts[0]) {
		case "bearer":
			ctx = context.WithValue(ctx, "auth_token", parts[1])
		case "basic":
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err == nil {
				ctx = context.WithValue(ctx, "auth_basic", string(decoded))
			}
		}
		return ctx
	}
}

// EncodeJSONResponse encodes response to JSON and sets the HTTP status code
// appropriately if response is a polykit.StandardResponse.
func EncodeJSONResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if resp, ok := response.(polykit.StandardResponse); ok {
		switch resp.ResponseCode {
		case "00":
			w.WriteHeader(http.StatusOK)
		case "01", "401":
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	return json.NewEncoder(w).Encode(response)
}

// EncodeFiberJSONResponse encodes response to JSON using GoFiber Context
// and sets the HTTP status code appropriately if response is a polykit.StandardResponse.
func EncodeFiberJSONResponse(ctx context.Context, c *fiber.Ctx, response interface{}) error {
	if resp, ok := response.(polykit.StandardResponse); ok {
		switch resp.ResponseCode {
		case "00":
			c.Status(fiber.StatusOK)
		case "01", "401":
			c.Status(fiber.StatusUnauthorized)
		default:
			c.Status(fiber.StatusBadRequest)
		}
	}
	return c.JSON(response)
}

