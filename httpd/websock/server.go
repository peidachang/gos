package websock

import (
	"fmt"
)

var pool = make(map[string]*Server, 0)

type SendMessages struct {
	ToClients []*Client
	Messages  []*Message
}

// Chat server.
type Server struct {
	clients map[int64]*Client
	addCh   chan *Client
	delCh   chan *Client
	sendCh  chan *SendMessages
	doneCh  chan bool
	errCh   chan error
}

func GetServer(name string) *Server {
	if v, ok := pool[name]; ok {
		return v
	}
	return nil
}

// Create new chat server.
func NewServer(name string) *Server {
	clients := make(map[int64]*Client)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	sendCh := make(chan *SendMessages)
	doneCh := make(chan bool)
	errCh := make(chan error)

	pool[name] = &Server{
		clients,
		addCh,
		delCh,
		sendCh,
		doneCh,
		errCh,
	}
	return pool[name]
}

func (s *Server) Add(c *Client) {
	s.addCh <- c
}

func (s *Server) Del(c *Client) {
	s.delCh <- c
}

func (s *Server) Done() {
	s.doneCh <- true
}

func (s *Server) Err(err error) {
	s.errCh <- err
}

func (s *Server) Send(c *Client, m *Message) {
	s.sendCh <- &SendMessages{[]*Client{c}, []*Message{m}}
}

func (s *Server) SendM(c []*Client, m []*Message) {
	s.sendCh <- &SendMessages{c, m}
}

func (s *Server) send(sm *SendMessages) {
	for _, c := range sm.ToClients {
		for _, m := range sm.Messages {
			c.Send(m)
		}
	}
}

// Listen and serve.
// It serves client connection and broadcast request.
func (s *Server) Start() {
	fmt.Println("Listening server...")

	for {
		select {

		// Add new a client
		case c := <-s.addCh:
			fmt.Println("Added new client")
			s.clients[c.id] = c
			// log.Println("Now", len(s.clients), "clients connected.")
			// s.sendPastMessages(c)

		// del a client
		case c := <-s.delCh:
			fmt.Println("Delete client")
			delete(s.clients, c.id)

		case smsg := <-s.sendCh:
			fmt.Println("Send msg:", smsg)
			s.send(smsg)

		case err := <-s.errCh:
			fmt.Println("Error:", err)

		case <-s.doneCh:
			return
		}
	}
}
