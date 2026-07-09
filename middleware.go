package polykit

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// LoggingMiddleware is an example of a core middleware that logs request execution time.
func LoggingMiddleware(logger *log.Logger) Middleware {
	return func(next Endpoint) Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Printf("request completed in %v, err: %v", time.Since(begin), err)
			}(time.Now())
			return next(ctx, request)
		}
	}
}

// JWTAuthMiddleware expects a token in the context under the key "auth_token".
// It validates the token and injects claims (e.g. user_id) back into the context.
func JWTAuthMiddleware(secretKey string) Middleware {
	return func(next Endpoint) Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			tokenString, ok := ctx.Value("auth_token").(string)
			if !ok || tokenString == "" {
				return StandardResponse{
					ResponseCode: "401",
					Message:      "Unauthorized: Missing token",
				}, nil
			}

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				return StandardResponse{
					ResponseCode: "401",
					Message:      "Unauthorized: Invalid token",
				}, nil
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return StandardResponse{
					ResponseCode: "401",
					Message:      "Unauthorized: Invalid claims format",
				}, nil
			}

			// Extract user_id or any other claim and pass to context
			if userID, ok := claims["user_id"]; ok {
				ctx = context.WithValue(ctx, "user_id", userID)
			} else {
				return StandardResponse{
					ResponseCode: "401",
					Message:      "Unauthorized: Missing user_id in claims",
				}, nil
			}

			// Call next endpoint
			return next(ctx, request)
		}
	}
}

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// Generic helper function to get value from context if needed
func GetContextValue(ctx context.Context, key ContextKey) interface{} {
	return ctx.Value(string(key))
}
