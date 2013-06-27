package httpd

import (
	"html/template"
	"io"
)

type IRender interface {
	render()
}

// EmptyRender
type EmptyRender struct{}

func (this *EmptyRender) render() {}

// TemplateRender
type TemplateRender struct {
	View   string
	Data   interface{}
	Writer io.Writer
}

func (this *TemplateRender) render() {
	tmpl, _ := template.ParseFiles("code/template" + Theme.GetTemplate() + this.View + ".htm")
	tmpl.Execute(this.Writer, this.Data)
}

// HeadRender
type HeadItemRender struct {
	Data   []string
	Writer io.Writer
}

func (this *HeadItemRender) render() {
	for _, v := range this.Data {
		this.Writer.Write([]byte(v + "\n"))
	}
}

// JsRender
type JsRender struct {
	Timestamp string
	Data      []string
	Writer    io.Writer
}

func (this *JsRender) render() {
	for _, v := range this.Data {
		this.Writer.Write([]byte("<script src=\"" + StaticUrl + Theme.GetJs() + v + "?ts=" + this.Timestamp + "\"></script>\n"))
	}
}

// CssRender
type CssRender struct {
	Timestamp string
	Data      []string
	Writer    io.Writer
}

func (this *CssRender) render() {
	for _, v := range this.Data {
		this.Writer.Write([]byte("<link href=\"" + StaticUrl + Theme.GetCss() + v + "?ts=\"" + this.Timestamp + " rel=\"stylesheet\"/>\n"))
	}
}

// TextRender
type TextRender struct {
	Name   string
	Source string
	Data   map[string]interface{}
	Writer io.Writer
}

func (this *TextRender) render() {

	tmpl, _ := template.New(this.Name).Parse(this.Source)
	tmpl.Execute(this.Writer, this.Data)
}
