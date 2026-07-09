package transport

import (
	"context"

	"github.com/addpur/go-polykit"
	"google.golang.org/grpc/metadata"
)

type GRPCDecodeRequestFunc func(context.Context, interface{}) (interface{}, error)

type GRPCEncodeResponseFunc func(context.Context, interface{}) (interface{}, error)

type GRPCRequestFunc func(ctx context.Context, md metadata.MD) context.Context

type GRPCResponseFunc func(ctx context.Context, header *metadata.MD, trailer *metadata.MD) context.Context

type GRPCServerOption func(*GRPCServer)

type GRPCServer struct {
	e      polykit.Endpoint
	dec    GRPCDecodeRequestFunc
	enc    GRPCEncodeResponseFunc
	before []GRPCRequestFunc
	after  []GRPCResponseFunc
}

func NewGRPCServer(
	e polykit.Endpoint,
	dec GRPCDecodeRequestFunc,
	enc GRPCEncodeResponseFunc,
	options ...GRPCServerOption,
) *GRPCServer {
	s := &GRPCServer{
		e:   e,
		dec: dec,
		enc: enc,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *GRPCServer) ServeGRPC(ctx context.Context, req interface{}) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	for _, f := range s.before {
		ctx = f(ctx, md)
	}

	request, err := s.dec(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := s.e(ctx, request)
	if err != nil {
		return nil, err
	}

	var mdHeader, mdTrailer metadata.MD
	for _, f := range s.after {
		ctx = f(ctx, &mdHeader, &mdTrailer)
	}

	return s.enc(ctx, response)
}

func GRPCServerBefore(before ...GRPCRequestFunc) GRPCServerOption {
	return func(s *GRPCServer) {
		s.before = append(s.before, before...)
	}
}

func GRPCServerAfter(after ...GRPCResponseFunc) GRPCServerOption {
	return func(s *GRPCServer) {
		s.after = append(s.after, after...)
	}
}
