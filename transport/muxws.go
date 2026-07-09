package transport

import (
	"context"
	"net/http"

	"github.com/addpur/go-polykit"
	"github.com/gorilla/websocket"
)

type MuxWSRequestFunc func(ctx context.Context, conn *websocket.Conn) context.Context

type MuxWSServerOption func(*MuxWSServer)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type MuxWSServer struct {
	e      polykit.Endpoint
	dec    WSDecodeRequestFunc
	enc    WSEncodeResponseFunc
	before []MuxWSRequestFunc
}

func NewMuxWSServer(
	e polykit.Endpoint,
	dec WSDecodeRequestFunc,
	enc WSEncodeResponseFunc,
	options ...MuxWSServerOption,
) *MuxWSServer {
	s := &MuxWSServer{
		e:   e,
		dec: dec,
		enc: enc,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *MuxWSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}

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

func MuxWSServerBefore(before ...MuxWSRequestFunc) MuxWSServerOption {
	return func(s *MuxWSServer) {
		s.before = append(s.before, before...)
	}
}
