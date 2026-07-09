package transport

import (
	"context"
	"strings"

	"github.com/addpur/go-polykit"
	"github.com/gofiber/fiber/v2"
)

// FiberDecodeRequestFunc extracts a user-domain request object from an HTTP request.
type FiberDecodeRequestFunc func(*fiber.Ctx) (interface{}, error)

// FiberEncodeResponseFunc encodes a user-domain response object into an HTTP response.
type FiberEncodeResponseFunc func(*fiber.Ctx, interface{}) error

// FiberEncodeErrorFunc encodes a user-domain error into an HTTP response.
type FiberEncodeErrorFunc func(*fiber.Ctx, error)

// NewFiberServer constructs a new fiber handler for the given endpoint.
func NewFiberServer(
	e polykit.Endpoint,
	dec FiberDecodeRequestFunc,
	enc FiberEncodeResponseFunc,
	errEnc FiberEncodeErrorFunc,
) fiber.Handler {
	if errEnc == nil {
		errEnc = DefaultFiberErrorEncoder
	}
	return func(c *fiber.Ctx) error {
		req, err := dec(c)
		if err != nil {
			errEnc(c, err)
			return nil
		}

		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}

		// Extract token from Authorization header (Bearer token)
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				ctx = context.WithValue(ctx, "auth_token", parts[1])
			} else {
				ctx = context.WithValue(ctx, "auth_token", authHeader)
			}
		}

		res, err := e(ctx, req)
		if err != nil {
			errEnc(c, err)
			return nil
		}

		if err := enc(c, res); err != nil {
			errEnc(c, err)
			return nil
		}

		return nil
	}
}

func DefaultFiberErrorEncoder(c *fiber.Ctx, err error) {
	c.Status(fiber.StatusInternalServerError).JSON(polykit.StandardResponse{
		ResponseCode: "500",
		Message:      err.Error(),
	})
}
