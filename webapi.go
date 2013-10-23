package httpd

import (
	"encoding/json"
	"github.com/jiorry/libs/log"
)

type ApiParams struct {
	Method string
	Args   interface{}
}

type WebApi struct {
	parent   interface{}
	UserAuth *UserAuth
	Ctx      *Context
}

func (w *WebApi) Prepare(ctx *Context, p interface{}) {
	w.Ctx = ctx
	w.parent = p
}

func (w *WebApi) Init() {}

func (w *WebApi) Reply(data interface{}, err error) {
	w.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.Ctx.WriteString(NewError(0, err).Json())
		return
	}

	encoder := json.NewEncoder(w.Ctx.ResponseWriter)
	if err := encoder.Encode(data); err != nil {
		log.App.Crit(err)
		return
	}
}
