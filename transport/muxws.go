package transport

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/addpur/go-polykit"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // WARNING: allow all origins for example purposes
	},
}

// NewMuxWSServer creates an http.HandlerFunc that upgrades the connection to WebSocket
// and wraps a polykit.Endpoint.
func NewMuxWSServer(
	e polykit.Endpoint,
	dec WSDecodeRequestFunc,
	enc WSEncodeResponseFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		ctx := r.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		// Get token from URL query string
		token := r.URL.Query().Get("token")
		if token != "" {
			ctx = context.WithValue(ctx, "auth_token", token)
		}

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}

			req, err := dec(mt, msg)
			if err != nil {
				continue
			}

			res, err := e(ctx, req)
			if err != nil {
				continue
			}

			outMt, outMsg, err := enc(res)
			if err != nil {
				continue
			}

			if err := c.WriteMessage(outMt, outMsg); err != nil {
				break
			}
		}
	}
}
