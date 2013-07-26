package httpd

import (
	"html/template"
	"io"
)

type IRender interface {
	Render(io.Writer)
}

// EmptyRender
type EmptyRender struct{}

func (this *EmptyRender) Render(w io.Writer) {}

// TemplateRender
type TemplateRender struct {
	View string
	Data interface{}
}

func (this *TemplateRender) Render(w io.Writer) {
	tmpl, _ := template.ParseFiles("code/template" + Theme.GetTemplate() + this.View + ".htm")
	tmpl.Execute(w, this.Data)
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
	Data []string
}

func (this *JsRender) Render(w io.Writer) {
	for _, v := range this.Data {
		w.Write([]byte("<script src=\"" + StaticUrl + Theme.GetJs() + v + httpServer.Timestamp + "\"></script>\n"))
	}
}

// CssRender
type CssRender struct {
	Data []string
}

func (this *CssRender) Render(w io.Writer) {
	for _, v := range this.Data {
		w.Write([]byte("<link href=\"" + StaticUrl + Theme.GetCss() + v + httpServer.Timestamp + "\" rel=\"stylesheet\"/>\n"))
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
