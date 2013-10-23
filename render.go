package httpd

import (
	"html/template"
	"io"
)

var (
	b_JS_TAG_BEGIN  = []byte("<script src=\"")
	b_JS_TAG_END    = []byte("\"></script>\n")
	b_CSS_TAG_BEGIN = []byte("<link href=\"")
	b_CSS_TAG_END   = []byte("\" rel=\"stylesheet\"/>\n")
)

type IRender interface {
	Render(io.Writer)
}

// EmptyRender
type EmptyRender struct{}

func (this *EmptyRender) Render(w io.Writer) {}

// TemplateRender
type TemplateRender struct {
	View *ThemeItem
	Data interface{}
}

func (this *TemplateRender) Render(w io.Writer) {
	defer func() {
		if err := recover(); err != nil {
			panic("template not found: " + httpServer.WebRoot + this.View.GetPath() + ".htm")
		}
	}()
	filepath := httpServer.WebRoot + this.View.GetPath() + ".htm"

	tmpl, _ := template.ParseFiles(filepath)
	err := tmpl.Execute(w, this.Data)
	if err != nil {
		panic("template execute error: " + filepath)
	}
}

// HeadRender
type HeadItemRender struct {
	Data []string
}

func (this *HeadItemRender) Render(w io.Writer) {
	for _, v := range this.Data {
		w.Write([]byte(v + "\n"))
	}
}

// JsRender
type JsRender struct {
	Data []*ThemeItem
}

func (this *JsRender) Render(w io.Writer) {
	for _, v := range this.Data {
		w.Write(b_JS_TAG_BEGIN)
		w.Write([]byte(StaticUrl + v.GetAssetsPath() + httpServer.Timestamp))
		w.Write(B_QUOTE)
		if v.Data != nil {
			for k, val := range v.Data {
				w.Write(B_SPACE)
				w.Write([]byte(k))
				w.Write(B_EQUAL)
				w.Write(B_QUOTE)
				w.Write([]byte(val))
				w.Write(B_QUOTE)
			}
		}
		w.Write(b_JS_TAG_END)
	}
}

// CssRender
type CssRender struct {
	Data []*ThemeItem
}

func (this *CssRender) Render(w io.Writer) {
	for _, v := range this.Data {
		w.Write(b_CSS_TAG_BEGIN)
		w.Write([]byte(StaticUrl + v.GetAssetsPath() + httpServer.Timestamp))
		w.Write(b_CSS_TAG_END)
	}
}

// TextRender
type TextRender struct {
	Name   string
	Source string
	Data   map[string]interface{}
}

func (this *TextRender) Render(w io.Writer) {

	tmpl, _ := template.New(this.Name).Parse(this.Source)
	tmpl.Execute(w, this.Data)
}
