package transport

import (
	"context"
	"strings"

	"github.com/addpur/go-polykit"
	"google.golang.org/grpc/metadata"
)

// GRPCDecodeRequestFunc extracts a user-domain request object from a gRPC request.
type GRPCDecodeRequestFunc func(context.Context, interface{}) (interface{}, error)

// GRPCEncodeResponseFunc encodes a user-domain response object into a gRPC response.
type GRPCEncodeResponseFunc func(context.Context, interface{}) (interface{}, error)

// GRPCServer wraps a polykit.Endpoint and implements a generic gRPC handler.
// Since gRPC generates typed interfaces, you typically embed this or call it
// from your generated server methods.
type GRPCServer struct {
	e   polykit.Endpoint
	dec GRPCDecodeRequestFunc
	enc GRPCEncodeResponseFunc
}

// NewGRPCServer constructs a new GRPCServer.
func NewGRPCServer(
	e polykit.Endpoint,
	dec GRPCDecodeRequestFunc,
	enc GRPCEncodeResponseFunc,
) *GRPCServer {
	return &GRPCServer{
		e:   e,
		dec: dec,
		enc: enc,
	}
}

// ServeGRPC handles a generic gRPC request and returns a generic gRPC response.
// This method should be called from your generated Protobuf service implementations.
func (s *GRPCServer) ServeGRPC(ctx context.Context, req interface{}) (interface{}, error) {
	// Extract metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		authTokens := md.Get("authorization")
		if len(authTokens) > 0 {
			authHeader := authTokens[0]
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				ctx = context.WithValue(ctx, "auth_token", parts[1])
			} else {
				ctx = context.WithValue(ctx, "auth_token", authHeader)
			}
		}
	}

	decodedReq, err := s.dec(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := s.e(ctx, decodedReq)
	if err != nil {
		return nil, err
	}

	return s.enc(ctx, res)
}
