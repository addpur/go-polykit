package transport

import (
	"context"

	"github.com/addpur/go-polykit"
	"github.com/gofiber/websocket/v2"
)

// WSDecodeRequestFunc extracts a request from a websocket message.
type WSDecodeRequestFunc func(messageType int, msg []byte) (interface{}, error)

// WSEncodeResponseFunc encodes a response into a websocket message.
type WSEncodeResponseFunc func(res interface{}) (messageType int, msg []byte, err error)

// NewFiberWSServer creates a WebSocket handler that wraps a polykit.Endpoint.
func NewFiberWSServer(
	e polykit.Endpoint,
	dec WSDecodeRequestFunc,
	enc WSEncodeResponseFunc,
) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		ctx := context.Background()

		// Get token from URL query string
		token := c.Query("token")
		if token != "" {
			ctx = context.WithValue(ctx, "auth_token", token)
		}

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break // Client disconnected or error
			}

			req, err := dec(mt, msg)
			if err != nil {
				continue // Skip invalid messages or handle error
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
				break // Write failed, close connection
			}
		}
	}
}
