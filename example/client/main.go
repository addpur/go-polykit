package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/addpur/go-polykit"
	"github.com/addpur/go-polykit/client"
)

func main() {
	ctx := context.Background()

	// ==========================================
	// 1. Fiber HTTP Client Example (FastHTTP)
	// ==========================================
	fiberClientEndpoint := client.NewFiberClient(
		"GET",
		"http://localhost:3000/fiber/secret",
		func(ctx context.Context, agent *fiber.Agent, request interface{}) error {
			// Set Header Authorization if needed
			agent.Request().Header.Set("Authorization", "Bearer MOCK_TOKEN_HERE")
			return nil // GET request doesn't need body encoding
		},
		func(ctx context.Context, agent *fiber.Agent, request interface{}) (interface{}, error) {
			// Extract body response
			var resp polykit.StandardResponse
			statusCode, body, errs := agent.Bytes()
			if len(errs) > 0 {
				return nil, errs[0]
			}
			log.Printf("[FiberClient] Status: %d, Response: %s", statusCode, string(body))
			return resp, nil
		},
	)

	// Invoke Fiber Client
	log.Println("--- Invoking Fiber HTTP Client ---")
	_, err := fiberClientEndpoint(ctx, nil)
	if err != nil {
		log.Printf("Fiber Client Error: %v\n", err)
	}

	// ==========================================
	// 2. Standard HTTP Client Example (net/http)
	// ==========================================
	// (Can be used for REST as well as GraphQL calls by changing the encoding to GraphQL JSON payload)
	stdClientEndpoint := client.NewHTTPClient(
		"GET",
		"http://localhost:8080/mux/hello?name=Polykit",
		func(ctx context.Context, req *http.Request, request interface{}) error {
			req.Header.Set("Authorization", "Bearer MOCK_TOKEN_HERE")
			return nil
		},
		func(ctx context.Context, resp *http.Response) (interface{}, error) {
			log.Printf("[StdClient] Status: %d", resp.StatusCode)
			return nil, nil
		},
		nil, // uses http.DefaultClient
	)

	log.Println("--- Invoking Standard HTTP Client ---")
	_, err = stdClientEndpoint(ctx, nil)
	if err != nil {
		log.Printf("Std Client Error: %v\n", err)
	}

	// ==========================================
	// 3. gRPC Client Example
	// ==========================================
	log.Println("--- Invoking gRPC Client ---")
	// Setup insecure connection for example purposes
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Assuming a proto service "SecretService" with method "GetSecret"
	type MockProtoResponse struct {
		Message string
	}
	grpcClientEndpoint := client.NewGRPCClient(
		conn,
		"SecretService",
		"GetSecret",
		func(ctx context.Context, request interface{}) (interface{}, error) {
			// Convert from Domain Request to Proto Request
			return request, nil
		},
		func(ctx context.Context, response interface{}) (interface{}, error) {
			// Convert from Proto Response to Domain Response
			return response, nil
		},
		&MockProtoResponse{}, // Pass pointer to expected proto reply struct
	)

	// Since we don't have a real gRPC server running on 50051, this will fail gracefully
	_, err = grpcClientEndpoint(ctx, "grpc-query")
	if err != nil {
		log.Printf("gRPC Client Error (Expected, as server is not running): %v\n", err)
	}
}
