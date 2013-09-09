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
	View *ThemeItem
	Data interface{}
}

func (this *TemplateRender) Render(w io.Writer) {
	defer func() {
		if err := recover(); err != nil {
			panic("template not found: " + httpServer.StaticDir + this.View.GetPath() + ".htm")
		}
	}()
	filepath := httpServer.StaticDir + this.View.GetPath() + ".htm"

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
	Theme *ThemeItem
	Data  []*ThemeItem
}

func (this *JsRender) Render(w io.Writer) {
	for _, v := range this.Data {
		w.Write([]byte("<script src=\"" + StaticUrl + v.GetAssetsPath() + httpServer.Timestamp + "\"></script>\n"))
	}
}

// CssRender
type CssRender struct {
	Theme *ThemeItem
	Data  []*ThemeItem
}

func (this *CssRender) Render(w io.Writer) {
	for _, v := range this.Data {
		w.Write([]byte("<link href=\"" + StaticUrl + v.GetAssetsPath() + httpServer.Timestamp + "\" rel=\"stylesheet\"/>\n"))
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
