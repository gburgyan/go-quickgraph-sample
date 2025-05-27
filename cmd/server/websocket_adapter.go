package main

import (
	"github.com/gburgyan/go-quickgraph"
	"github.com/gorilla/websocket"
	"net/http"
)

// GorillaWebSocketAdapter implements quickgraph.SimpleWebSocketConn
type GorillaWebSocketAdapter struct {
	conn *websocket.Conn
}

func (a *GorillaWebSocketAdapter) ReadMessage() ([]byte, error) {
	_, data, err := a.conn.ReadMessage()
	return data, err
}

func (a *GorillaWebSocketAdapter) WriteMessage(data []byte) error {
	return a.conn.WriteMessage(websocket.TextMessage, data)
}

func (a *GorillaWebSocketAdapter) Close() error {
	return a.conn.Close()
}

// GorillaWebSocketUpgrader implements the WebSocket upgrader interface
type GorillaWebSocketUpgrader struct {
	upgrader websocket.Upgrader
}

func NewGorillaUpgrader() *GorillaWebSocketUpgrader {
	return &GorillaWebSocketUpgrader{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, configure this properly
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    []string{"graphql-transport-ws"},
		},
	}
}

func (u *GorillaWebSocketUpgrader) Upgrade(w http.ResponseWriter, r *http.Request) (quickgraph.SimpleWebSocketConn, error) {
	// Set the response header to include the negotiated subprotocol
	responseHeader := http.Header{}
	if websocket.Subprotocols(r) != nil {
		for _, requestedProtocol := range websocket.Subprotocols(r) {
			if requestedProtocol == "graphql-transport-ws" {
				responseHeader.Set("Sec-WebSocket-Protocol", "graphql-transport-ws")
				break
			}
		}
	}

	conn, err := u.upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return &GorillaWebSocketAdapter{conn: conn}, nil
}
