package polykit

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)


// BasicAuthMiddleware checks credentials from the context key "auth_basic" (set by transport layer).
// The value is expected to be in the format "username:password" (decoded from Basic base64 header).
// If the credentials do not match, it returns a StandardResponse with ResponseCode "01".
func BasicAuthMiddleware(expectedUsername, expectedPassword string) Middleware {
	return func(next Endpoint) Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			authBasic, ok := ctx.Value("auth_basic").(string)
			if !ok || authBasic == "" {
				return StandardResponse{
					ResponseCode: "01",
					Message:      "Unauthorized: Missing Basic credentials",
				}, nil
			}

			parts := strings.SplitN(authBasic, ":", 2)
			if len(parts) != 2 || parts[0] != expectedUsername || parts[1] != expectedPassword {
				return StandardResponse{
					ResponseCode: "01",
					Message:      "Unauthorized: Invalid credentials",
				}, nil
			}

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

			if userID, ok := claims["user_id"]; ok {
				ctx = context.WithValue(ctx, "user_id", userID)
			} else {
				return StandardResponse{
					ResponseCode: "401",
					Message:      "Unauthorized: Missing user_id in claims",
				}, nil
			}

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
