package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/addpur/go-polykit"
	"github.com/addpur/go-polykit/transport"
)

// 1. Define your domain request and response
type SecretRequest struct {
	Query string `json:"query"`
}

// 2. Write your core business logic as an Endpoint
func makeSecretEndpoint() polykit.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SecretRequest)
		
		// Extract user_id injected by JWTAuthMiddleware
		userID := ctx.Value("user_id")
		
		return polykit.StandardResponse{
			ResponseCode: "00",
			Message:      "Success",
			Data: fmt.Sprintf("Hello User %v, your secret query '%s' has been processed.", userID, req.Query),
		}, nil
	}
}

func main() {
	secretKey := "my-very-secret-key"

	// Initialize the endpoint and wrap with middleware
	endpoint := makeSecretEndpoint()
	endpoint = polykit.LoggingMiddleware(log.Default())(endpoint)
	endpoint = polykit.JWTAuthMiddleware(secretKey)(endpoint)

	// ==========================================
	// Example 1: Serve via GoFiber (HTTP)
	// ==========================================
	app := fiber.New()
	fiberHandler := transport.NewFiberServer(
		endpoint,
		func(c *fiber.Ctx) (interface{}, error) {
			var req SecretRequest
			if err := c.QueryParser(&req); err != nil {
				return nil, err
			}
			return req, nil
		},
		func(c *fiber.Ctx, response interface{}) error {
			return c.JSON(response)
		},
		nil,
	)
	app.Get("/fiber/secret", fiberHandler)

	// ==========================================
	// Example 2: Serve via GoFiber (WebSocket)
	// ==========================================
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	wsHandler := transport.NewFiberWSServer(
		endpoint,
		func(mt int, msg []byte) (interface{}, error) {
			// For simplicity, assuming message is plain text query
			return SecretRequest{Query: string(msg)}, nil
		},
		func(res interface{}) (int, []byte, error) {
			// Basic formatting for example
			respStr := fmt.Sprintf("%v", res)
			return websocket.TextMessage, []byte(respStr), nil
		},
	)
	app.Get("/ws/secret", websocket.New(wsHandler))

	go func() {
		log.Println("Starting Fiber server (HTTP & WS) on :3000")
		if err := app.Listen(":3000"); err != nil {
			log.Fatalf("Fiber server error: %v", err)
		}
	}()

	// ==========================================
	// Example 3: Serve via gRPC (Conceptual Wrapper)
	// ==========================================
	// This shows how the gRPC handler would be initialized
	grpcHandler := transport.NewGRPCServer(
		endpoint,
		func(ctx context.Context, req interface{}) (interface{}, error) {
			// Cast from proto request to domain request
			return SecretRequest{Query: "grpc-query"}, nil
		},
		func(ctx context.Context, res interface{}) (interface{}, error) {
			// Cast from domain response to proto response
			return res, nil 
		},
	)
	log.Printf("gRPC handler initialized (requires grpc.Server and proto registration to actually run): %v", grpcHandler != nil)

	// Keep main running (or use graceful shutdown)
	select {}
}
