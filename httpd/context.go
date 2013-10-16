package httpd

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

type Context struct {
	ResponseWriter http.ResponseWriter
	routerParams   map[string]string
	Request        *http.Request
}

func (ctx *Context) WriteString(content string) {
	ctx.ResponseWriter.Write([]byte(content))
}

func (ctx *Context) Exit(code int, body string) {
	ctx.ResponseWriter.WriteHeader(code)
	ctx.ResponseWriter.Write([]byte(body))
}

func (ctx *Context) Redirect(urlStr string, args ...interface{}) {
	// ctx.ResponseWriter.Header().Set("Location", urlStr)
	// ctx.ResponseWriter.WriteHeader(302)

	http.Redirect(ctx.ResponseWriter, ctx.Request, fmt.Sprintf(urlStr, args...), http.StatusSeeOther)
}

func (ctx *Context) NotModified() {
	ctx.ResponseWriter.WriteHeader(304)
}

func (ctx *Context) NotFound(message string) {
	ctx.ResponseWriter.WriteHeader(404)
	ctx.ResponseWriter.Write([]byte(message))
}

func (ctx *Context) ContentType(ext string) {
	ctype := mime.TypeByExtension(ext)

	if ctype != "" {
		ctx.ResponseWriter.Header().Set("Content-Type", ctype)
	}
}

func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
	if unique {
		ctx.ResponseWriter.Header().Set(hdr, val)
	} else {
		ctx.ResponseWriter.Header().Add(hdr, val)
	}
}

func (ctx *Context) SetCookie(name string, value string, age int64, path string, domain string, httpOnly bool) {
	var expires time.Time
	if age != 0 {
		expires = time.Unix(time.Now().Unix()+age, 0)
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   domain,
		Expires:  expires,
		HttpOnly: httpOnly,
	}
	http.SetCookie(ctx.ResponseWriter, cookie)
}

func (ctx *Context) CheckLogin(loginUrl string) bool {
	if auth := (&UserAuth{}).SetContext(ctx); auth.NotOk() {
		auth.ClearCookie()
		ctx.Redirect(loginUrl + "?redirect=" + ctx.Request.URL.RequestURI())
		return false
	}
	return true
}

func (ctx *Context) RouterParam(key string) string {
	if v, ok := ctx.routerParams[key]; ok {
		return v
	} else {
		return ""
	}
}

func webTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

type FileContext struct {
	Writer io.Writer
}
