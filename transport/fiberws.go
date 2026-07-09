package transport

import (
	"context"

	"github.com/addpur/go-polykit"
	"github.com/gofiber/websocket/v2"
)

type WSDecodeRequestFunc func(messageType int, msg []byte) (interface{}, error)

type WSEncodeResponseFunc func(res interface{}) (messageType int, msg []byte, err error)

type WSRequestFunc func(ctx context.Context, conn *websocket.Conn) context.Context

type WSServerOption func(*FiberWSServer)

type FiberWSServer struct {
	e      polykit.Endpoint
	dec    WSDecodeRequestFunc
	enc    WSEncodeResponseFunc
	before []WSRequestFunc
}

func NewFiberWSServer(
	e polykit.Endpoint,
	dec WSDecodeRequestFunc,
	enc WSEncodeResponseFunc,
	options ...WSServerOption,
) *FiberWSServer {
	s := &FiberWSServer{
		e:   e,
		dec: dec,
		enc: enc,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *FiberWSServer) Handler() func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		ctx := context.Background()

		for _, f := range s.before {
			ctx = f(ctx, c)
		}

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}

			req, err := s.dec(mt, msg)
			if err != nil {
				continue
			}

			res, err := s.e(ctx, req)
			if err != nil {
				continue
			}

			outMt, outMsg, err := s.enc(res)
			if err != nil {
				continue
			}

			if err := c.WriteMessage(outMt, outMsg); err != nil {
				break
			}
		}
	}
}

func WSServerBefore(before ...WSRequestFunc) WSServerOption {
	return func(s *FiberWSServer) {
		s.before = append(s.before, before...)
	}
}
