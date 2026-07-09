package transport

import (
	"context"

	"github.com/addpur/go-polykit"
	"github.com/gofiber/fiber/v2"
)

type FiberDecodeRequestFunc func(context.Context, *fiber.Ctx) (interface{}, error)

type FiberEncodeResponseFunc func(context.Context, *fiber.Ctx, interface{}) error

type FiberEncodeErrorFunc func(context.Context, error, *fiber.Ctx)

type FiberRequestFunc func(ctx context.Context, c *fiber.Ctx) context.Context

type FiberResponseFunc func(ctx context.Context, c *fiber.Ctx) context.Context

type FiberServerOption func(*FiberServer)

type FiberServer struct {
	e      polykit.Endpoint
	dec    FiberDecodeRequestFunc
	enc    FiberEncodeResponseFunc
	errEnc FiberEncodeErrorFunc
	before []FiberRequestFunc
	after  []FiberResponseFunc
}

func NewFiberServer(
	e polykit.Endpoint,
	dec FiberDecodeRequestFunc,
	enc FiberEncodeResponseFunc,
	options ...FiberServerOption,
) *FiberServer {
	s := &FiberServer{
		e:      e,
		dec:    dec,
		enc:    enc,
		errEnc: DefaultFiberErrorEncoder,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *FiberServer) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}

		for _, f := range s.before {
			ctx = f(ctx, c)
		}

		request, err := s.dec(ctx, c)
		if err != nil {
			s.errEnc(ctx, err, c)
			return nil
		}

		response, err := s.e(ctx, request)
		if err != nil {
			s.errEnc(ctx, err, c)
			return nil
		}

		for _, f := range s.after {
			ctx = f(ctx, c)
		}

		if err := s.enc(ctx, c, response); err != nil {
			s.errEnc(ctx, err, c)
			return nil
		}

		return nil
	}
}

func FiberServerBefore(before ...FiberRequestFunc) FiberServerOption {
	return func(s *FiberServer) {
		s.before = append(s.before, before...)
	}
}

func FiberServerAfter(after ...FiberResponseFunc) FiberServerOption {
	return func(s *FiberServer) {
		s.after = append(s.after, after...)
	}
}

func FiberServerErrorEncoder(errEnc FiberEncodeErrorFunc) FiberServerOption {
	return func(s *FiberServer) {
		s.errEnc = errEnc
	}
}

func DefaultFiberErrorEncoder(ctx context.Context, err error, c *fiber.Ctx) {
	c.Status(fiber.StatusInternalServerError).JSON(polykit.StandardResponse{
		ResponseCode: "500",
		Message:      err.Error(),
	})
}
