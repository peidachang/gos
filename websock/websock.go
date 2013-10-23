package websock

import (
	"code.google.com/p/go.net/websocket"
)

type WebSock struct {
	Ws      *websocket.Conn
	Control IControl
	Server  *Server
}

func (w *WebSock) Prepare(ws *websocket.Conn, control IControl, s *Server) {
	w.Ws = ws
	w.Control = control
	w.Server = s
}

func (w *WebSock) Init() {}

func (w *WebSock) Listen() {
	defer func() {
		err := w.Ws.Close()
		if err != nil {
			w.Server.errCh <- err
		}
	}()

	c := NewClient(w.Control.ClientId(), w.Ws, w.Control, w.Server)
	w.Server.Add(c)
	c.Listen()
}

var maxId int64 = 0

func (w *WebSock) ClientId() int64 {
	maxId++
	return maxId
}

func (w *WebSock) Receive(c *Client, msg *Message) {
	println("websock")
}
