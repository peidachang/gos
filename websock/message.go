package websock

import (
	"fmt"
)

type Message struct {
	Method string      `json: "method"`
	Args   interface{} `json:"args"`
	IType  int         `json:"itype"`
}

func (m *Message) String() string {
	return fmt.Sprintln(m.Method, m.Args, m.IType)
}
