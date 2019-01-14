package gorillaconnection

import (
	"encoding/base64"
	"github.com/gorilla/websocket"
)

type Connection struct {
	c *websocket.Conn
}

func (g *Connection) Receive() ([]byte, error) {
	_, p, err := g.c.ReadMessage()
	if err != nil {
		return nil, err
	}
	bytes, err := base64.StdEncoding.DecodeString(string(p))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (g *Connection) Send(b []byte) error {
	str := base64.StdEncoding.EncodeToString(b)
	return g.c.WriteMessage(websocket.TextMessage, []byte(str))
}
