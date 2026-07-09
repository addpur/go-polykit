package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/addpur/go-polykit"
	"github.com/addpur/go-polykit/client"
)

func basicAuthHeader(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func main() {
	ctx := context.Background()

	fiberClientEndpoint := client.NewFiberClient(
		"GET",
		"http://localhost:3000/fiber/basic-secret?query=hello-from-fiber-client",
		func(ctx context.Context, agent *fiber.Agent, request interface{}) error {
			agent.Request().Header.Set("Authorization", basicAuthHeader("admin", "s3cr3t"))
			return nil
		},
		func(ctx context.Context, statusCode int, body []byte) (interface{}, error) {
			var resp polykit.StandardResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				return nil, fmt.Errorf("decode error: %w", err)
			}
			log.Printf("[FiberClient] Status: %d, ResponseCode: %s, Data: %v", statusCode, resp.ResponseCode, resp.Data)
			return resp, nil
		},
	)

	log.Println("--- Invoking Fiber HTTP Client (Basic Auth) ---")
	_, err := fiberClientEndpoint(ctx, nil)
	if err != nil {
		log.Printf("Fiber Client Error: %v\n", err)
	}

	stdClientEndpoint := client.NewHTTPClient(
		"GET",
		"http://localhost:8080/mux/basic-secret?query=hello-from-std-client",
		func(ctx context.Context, req *http.Request, request interface{}) error {
			req.Header.Set("Authorization", basicAuthHeader("admin", "s3cr3t"))
			return nil
		},
		func(ctx context.Context, resp *http.Response) (interface{}, error) {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			var stdResp polykit.StandardResponse
			if err := json.Unmarshal(body, &stdResp); err != nil {
				return nil, fmt.Errorf("decode error: %w", err)
			}
			log.Printf("[StdClient] Status: %d, ResponseCode: %s, Data: %v", resp.StatusCode, stdResp.ResponseCode, stdResp.Data)
			return stdResp, nil
		},
		nil,
	)

	log.Println("--- Invoking Standard HTTP Client (Basic Auth) ---")
	_, err = stdClientEndpoint(ctx, nil)
	if err != nil {
		log.Printf("Std Client Error: %v\n", err)
	}

	log.Println("--- Invoking gRPC Client (no server running, graceful fail expected) ---")
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	type MockProtoResponse struct {
		Message string
	}
	grpcClientEndpoint := client.NewGRPCClient(
		conn,
		"SecretService",
		"GetSecret",
		func(ctx context.Context, request interface{}) (interface{}, error) {
			return request, nil
		},
		func(ctx context.Context, response interface{}) (interface{}, error) {
			return response, nil
		},
		&MockProtoResponse{},
	)

	_, err = grpcClientEndpoint(ctx, "grpc-query")
	if err != nil {
		log.Printf("gRPC Client Error (expected): %v\n", err)
	}
}
