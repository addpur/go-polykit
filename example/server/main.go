package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/addpur/go-polykit"
	"github.com/addpur/go-polykit/pkg/polykit/logger"
	"github.com/addpur/go-polykit/pkg/polykit/telemetry"
	"github.com/addpur/go-polykit/transport"
)

type SecretRequest struct {
	Query string `json:"query"`
}

func makeJWTEndpoint() polykit.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SecretRequest)
		userID := ctx.Value("user_id")
		return polykit.StandardResponse{
			ResponseCode: "00",
			Message:      "Success",
			Data:         fmt.Sprintf("Hello User %v, your query: '%s'", userID, req.Query),
		}, nil
	}
}

func makeBasicAuthEndpoint() polykit.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SecretRequest)
		return polykit.StandardResponse{
			ResponseCode: "00",
			Message:      "Success",
			Data:         fmt.Sprintf("Authenticated! Your query: '%s'", req.Query),
		}, nil
	}
}

func fiberDecodeQuery(ctx context.Context, c *fiber.Ctx) (interface{}, error) {
	var req SecretRequest
	if err := c.QueryParser(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func fiberEncodeJSON(ctx context.Context, c *fiber.Ctx, response interface{}) error {
	return c.JSON(response)
}

func fiberErrorEncoder(ctx context.Context, err error, c *fiber.Ctx) {
	c.Status(fiber.StatusInternalServerError).JSON(polykit.StandardResponse{
		ResponseCode: "99",
		Message:      err.Error(),
	})
}

func muxDecodeQuery(ctx context.Context, r *http.Request) (interface{}, error) {
	return SecretRequest{Query: r.URL.Query().Get("query")}, nil
}

func muxEncodeJSON(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if resp, ok := response.(polykit.StandardResponse); ok && resp.ResponseCode != "00" {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return json.NewEncoder(w).Encode(response)
}

func muxErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(polykit.StandardResponse{
		ResponseCode: "99",
		Message:      err.Error(),
	})
}

func main() {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize Zap: %v", err)
	}
	defer zapLogger.Sync()
	sugar := zapLogger.Sugar()
	zapLog := logger.NewLogger(sugar)

	jwtSecretKey := "my-very-secret-key"
	basicUser := "admin"
	basicPass := "s3cr3t"

	jwtEndpoint := polykit.Chain(
		telemetry.TracingMiddleware("jwt-secret-endpoint"),
		logger.LoggingMiddleware(zapLog, "jwt-secret-endpoint"),
		polykit.JWTAuthMiddleware(jwtSecretKey),
	)(makeJWTEndpoint())

	basicEndpoint := polykit.Chain(
		telemetry.TracingMiddleware("basic-secret-endpoint"),
		logger.LoggingMiddleware(zapLog, "basic-secret-endpoint"),
		polykit.BasicAuthMiddleware(basicUser, basicPass),
	)(makeBasicAuthEndpoint())

	// Fiber ServerOption — mirip go-kit kithttp.ServerOption
	fiberOpt := []transport.FiberServerOption{
		transport.FiberServerErrorEncoder(fiberErrorEncoder),
		transport.FiberServerBefore(
			transport.FiberToContext(),
		),
		transport.FiberServerAfter(
			transport.SetFiberResponseHeader("X-Frame-Options", "DENY"),
			transport.SetFiberResponseHeader("X-Content-Type-Options", "nosniff"),
		),
	}

	app := fiber.New()

	app.Get("/fiber/jwt-secret", transport.NewFiberServer(
		jwtEndpoint, fiberDecodeQuery, fiberEncodeJSON, fiberOpt...,
	).Handler())

	app.Get("/fiber/basic-secret", transport.NewFiberServer(
		basicEndpoint, fiberDecodeQuery, fiberEncodeJSON, fiberOpt...,
	).Handler())

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/secret", websocket.New(transport.NewFiberWSServer(
		jwtEndpoint,
		func(mt int, msg []byte) (interface{}, error) {
			return SecretRequest{Query: string(msg)}, nil
		},
		func(res interface{}) (int, []byte, error) {
			return websocket.TextMessage, []byte(fmt.Sprintf("%v", res)), nil
		},
	).Handler()))

	go func() {
		sugar.Info("Starting Fiber server on :3000")
		if err := app.Listen(":3000"); err != nil {
			sugar.Fatalf("Fiber server error: %v", err)
		}
	}()

	// Mux ServerOption — mirip go-kit kithttp.ServerOption
	muxOpt := []transport.HTTPServerOption{
		transport.HTTPServerErrorEncoder(muxErrorEncoder),
		transport.HTTPServerBefore(
			transport.HTTPToContext(),
			transport.PopulateRequestID(),
		),
		transport.HTTPServerAfter(
			transport.SetResponseHeader("X-Frame-Options", "DENY"),
			transport.SetResponseHeader("X-Content-Type-Options", "nosniff"),
			transport.SetResponseHeader("X-XSS-Protection", "1; mode=block"),
		),
	}

	r := mux.NewRouter()

	r.Methods(http.MethodGet).Path("/mux/jwt-secret").Handler(
		transport.NewHTTPServer(jwtEndpoint, muxDecodeQuery, muxEncodeJSON, muxOpt...),
	)

	r.Methods(http.MethodGet).Path("/mux/basic-secret").Handler(
		transport.NewHTTPServer(basicEndpoint, muxDecodeQuery, muxEncodeJSON, muxOpt...),
	)

	go func() {
		sugar.Info("Starting Gorilla Mux server on :8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			sugar.Fatalf("Mux server error: %v", err)
		}
	}()

	select {}
}
