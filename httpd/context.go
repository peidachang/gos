package httpd

import (
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

type Context struct {
	ResponseWriter http.ResponseWriter
	RouterParams   map[string]string
	Request        *http.Request
}

func (ctx *Context) WriteString(content string) {
	ctx.ResponseWriter.Write([]byte(content))
}

func (ctx *Context) Exit(code int, body string) {
	ctx.ResponseWriter.WriteHeader(code)
	ctx.ResponseWriter.Write([]byte(body))
}

func (ctx *Context) Redirect(urlStr string) {
	// ctx.ResponseWriter.Header().Set("Location", urlStr)
	// ctx.ResponseWriter.WriteHeader(302)

	http.Redirect(ctx.ResponseWriter, ctx.Request, urlStr, http.StatusSeeOther)
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

func (ctx *Context) CheckLogin(path string) bool {
	if auth := (&UserAuth{}).SetContext(ctx); auth.NotOk() {
		auth.ClearCookie()
		ctx.Redirect(path + "?redirect=" + ctx.Request.URL.RequestURI())
		return false
	}
	return true
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
