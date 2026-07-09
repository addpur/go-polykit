package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/addpur/go-polykit"
)

// GQLDecodeRequestFunc extracts a user-domain request from GraphQL ResolveParams.
type GQLDecodeRequestFunc func(graphql.ResolveParams) (interface{}, error)

// GQLEncodeResponseFunc encodes a user-domain response into a format suitable for GraphQL.
type GQLEncodeResponseFunc func(interface{}) (interface{}, error)

// NewGQLResolver creates a graphql-go resolver function that wraps a polykit.Endpoint.
func NewGQLResolver(
	e polykit.Endpoint,
	dec GQLDecodeRequestFunc,
	enc GQLEncodeResponseFunc,
) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		req, err := dec(p)
		if err != nil {
			return nil, err
		}

		ctx := p.Context
		if ctx == nil {
			ctx = context.Background()
		}

		// Try to extract from standard HTTP request if present in context (graphql-go/handler injects this)
		if reqObj, ok := ctx.Value("request").(*http.Request); ok && reqObj != nil {
			authHeader := reqObj.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					ctx = context.WithValue(ctx, "auth_token", parts[1])
				} else {
					ctx = context.WithValue(ctx, "auth_token", authHeader)
				}
			}
		}

		res, err := e(ctx, req)
		if err != nil {
			return nil, err
		}

		return enc(res)
	}
}
