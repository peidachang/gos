package httpd

import (
	"html/template"
	"io"
)

type IRender interface {
	render(io.Writer)
}

// EmptyRender
type EmptyRender struct{}

func (this *EmptyRender) render(w io.Writer) {}

// TemplateRender
type TemplateRender struct {
	View string
	Data interface{}
}

func (this *TemplateRender) render(w io.Writer) {
	tmpl, _ := template.ParseFiles("code/template" + Theme.GetTemplate() + this.View + ".htm")
	tmpl.Execute(w, this.Data)
}

// HeadRender
type HeadItemRender struct {
	Data []string
}

func (this *HeadItemRender) render(w io.Writer) {
	for _, v := range this.Data {
		w.Write([]byte(v + "\n"))
	}
}

// JsRender
type JsRender struct {
	Timestamp string
	Data      []string
}

func (this *JsRender) render(w io.Writer) {
	for _, v := range this.Data {
		w.Write([]byte("<script src=\"" + StaticUrl + Theme.GetJs() + v + "?ts=" + this.Timestamp + "\"></script>\n"))
	}
}

// CssRender
type CssRender struct {
	Timestamp string
	Data      []string
}

func (this *CssRender) render(w io.Writer) {
	for _, v := range this.Data {
		w.Write([]byte("<link href=\"" + StaticUrl + Theme.GetCss() + v + "?ts=\"" + this.Timestamp + " rel=\"stylesheet\"/>\n"))
	}
}

// TextRender
type TextRender struct {
	Name   string
	Source string
	Data   map[string]interface{}
}

func (this *TextRender) render(w io.Writer) {

	tmpl, _ := template.New(this.Name).Parse(this.Source)
	tmpl.Execute(w, this.Data)
}
