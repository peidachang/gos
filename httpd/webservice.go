package httpd

import (
	"encoding/json"
	"github.com/jiorry/gos/log"
)

type WSParams struct {
	Method string
	Args   interface{}
}

// func (this *WSParams) GetMethod() string {
// 	return this.method
// }

// func (this *WSParams) GetArgs() map[string]interface{} {
// 	return this.args
// }

type WebService struct {
	Ctx *Context
}

func (this *WebService) SetContext(ctx *Context) {
	this.Ctx = ctx
}

func (this *WebService) Reply(data interface{}) {
	this.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(this.Ctx.ResponseWriter)
	if err := encoder.Encode(data); err != nil {
		log.App.Crit(err)
		return
	}
}
