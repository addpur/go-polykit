package client

import (
	"context"

	"github.com/addpur/go-polykit"
	"google.golang.org/grpc"
)

// GRPCEncodeRequestFunc encodes a user-domain request into a gRPC request.
type GRPCEncodeRequestFunc func(context.Context, interface{}) (interface{}, error)

// GRPCDecodeResponseFunc extracts a user-domain response from a gRPC response.
type GRPCDecodeResponseFunc func(context.Context, interface{}) (interface{}, error)

// NewGRPCClient creates a polykit.Endpoint that makes a gRPC call.
func NewGRPCClient(
	conn *grpc.ClientConn,
	serviceName string,
	methodName string,
	enc GRPCEncodeRequestFunc,
	dec GRPCDecodeResponseFunc,
	replyType interface{}, // The generated protobuf response struct pointer
) polykit.Endpoint {
	fullMethod := "/" + serviceName + "/" + methodName

	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, err := enc(ctx, request)
		if err != nil {
			return nil, err
		}

		// replyType is a dummy object just to pass the type to grpc.Invoke,
		// in real-world you would probably use reflection to create a new instance,
		// or rely on the decode func to handle the raw message if using a more dynamic approach.
		// For simplicity without reflection, we expect reply to be provided properly or handled by caller.

		err = conn.Invoke(ctx, fullMethod, req, replyType)
		if err != nil {
			return nil, err
		}

		return dec(ctx, replyType)
	}
}
