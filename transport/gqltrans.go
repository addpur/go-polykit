package transport

import (
	"context"

	"github.com/addpur/go-polykit"
	"github.com/graphql-go/graphql"
)

type GQLDecodeRequestFunc func(context.Context, graphql.ResolveParams) (interface{}, error)

type GQLEncodeResponseFunc func(context.Context, interface{}) (interface{}, error)

type GQLRequestFunc func(ctx context.Context, p graphql.ResolveParams) context.Context

type GQLServerOption func(*GQLServer)

type GQLServer struct {
	e      polykit.Endpoint
	dec    GQLDecodeRequestFunc
	enc    GQLEncodeResponseFunc
	before []GQLRequestFunc
}

func NewGQLResolver(
	e polykit.Endpoint,
	dec GQLDecodeRequestFunc,
	enc GQLEncodeResponseFunc,
	options ...GQLServerOption,
) *GQLServer {
	s := &GQLServer{
		e:   e,
		dec: dec,
		enc: enc,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *GQLServer) Resolve() graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		ctx := p.Context
		if ctx == nil {
			ctx = context.Background()
		}

		for _, f := range s.before {
			ctx = f(ctx, p)
		}

		request, err := s.dec(ctx, p)
		if err != nil {
			return nil, err
		}

		response, err := s.e(ctx, request)
		if err != nil {
			return nil, err
		}

		return s.enc(ctx, response)
	}
}

func GQLServerBefore(before ...GQLRequestFunc) GQLServerOption {
	return func(s *GQLServer) {
		s.before = append(s.before, before...)
	}
}
