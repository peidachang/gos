package httpd

import (
	"encoding/json"
	"github.com/jiorry/gos/log"
)

type WSParams struct {
	Method string
	Args   interface{}
}

type WebService struct {
	UserAuth        *UserAuth
	Ctx             *Context
	RequireAuth     bool
	PublicFunctions []string
}

func (this *WebService) SetContext(ctx *Context) {
	this.Ctx = ctx
}

func (this *WebService) Init() {}

func (this *WebService) IsAuth() bool {
	this.UserAuth = (&UserAuth{}).SetContext(this.Ctx)
	return this.UserAuth.IsOk()
}

func (this *WebService) Reply(data interface{}, err error) {
	this.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	if err != nil {
		this.Ctx.WriteString(MyErr(0, err).Json())
		return
	}

	encoder := json.NewEncoder(this.Ctx.ResponseWriter)
	if err := encoder.Encode(data); err != nil {
		log.App.Crit(err)
		return
	}
}
