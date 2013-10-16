package websock

import (
	"fmt"
	"io"

	"code.google.com/p/go.net/websocket"
)

type IControl interface {
	Receive(*Client, *Message)
	ClientId() int64
}

// Chat client.
type Client struct {
	id      int64
	ws      *websocket.Conn
	server  *Server
	ch      chan *Message
	doneCh  chan bool
	control IControl
}

// Create new chat client.
func NewClient(cid int64, ws *websocket.Conn, control IControl, server *Server) *Client {
	if ws == nil {
		panic("ws cannot be nil")
	}

	if server == nil {
		panic("server cannot be nil")
	}

	ch := make(chan *Message)
	doneCh := make(chan bool)

	return &Client{cid, ws, server, ch, doneCh, control}
}

func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

func (c *Client) Send(msg *Message) {
	select {
	case c.ch <- msg:
	default:
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected.", c.id)
		c.server.Err(err)
	}
}

func (c *Client) Done() {
	c.doneCh <- true
}

// Listen Write and Read request via chanel
func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

// Listen write request via chanel
func (c *Client) listenWrite() {
	fmt.Println("Listening write to client")
	for {
		select {

		// send message to the client
		case msg := <-c.ch:
			fmt.Println("Send:", msg)
			websocket.JSON.Send(c.ws, msg)

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true // for listenRead method
			return
		}
	}
}

// Listen read request via chanel
func (c *Client) listenRead() {
	fmt.Println("Listening read from client")
	for {
		select {

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true // for listenWrite method
			return

		// read data from websocket connection
		default:
			var msg Message
			err := websocket.JSON.Receive(c.ws, &msg)
			if err == io.EOF {
				c.doneCh <- true
			} else if err != nil {
				c.server.Err(err)
			} else {
				// c.server.Send(c, &msg)
				fmt.Println("receive: ", msg)
				c.control.Receive(c, &msg)
			}
		}
	}
}
