package httpd

import (
	"io"
)

type AppLayout struct {
	headLayout    *HeadLayout
	topRender     IRender
	headerRender  IRender
	contextRender IRender
	footerRender  IRender
	bottomRender  IRender
}

func (this *AppLayout) TopView(theme string, name string, data interface{}) {
	this.topRender = &TemplateRender{
		View: &ThemeItem{theme, "template", name},
		Data: data}
}
func (this *AppLayout) HeaderView(theme string, name string, data interface{}) {
	this.headerRender = &TemplateRender{
		View: &ThemeItem{theme, "template", name},
		Data: data}
}
func (this *AppLayout) FooterView(theme string, name string, data interface{}) {
	this.footerRender = &TemplateRender{
		View: &ThemeItem{theme, "template", name},
		Data: data}
}
func (this *AppLayout) BottomView(theme string, name string, data interface{}) {
	this.bottomRender = &TemplateRender{
		View: &ThemeItem{theme, "template", name},
		Data: data}
}
func (this *AppLayout) SetTopRender(r IRender) {
	this.contextRender = r
}
func (this *AppLayout) SetHeaderRender(r IRender) {
	this.contextRender = r
}
func (this *AppLayout) SetContextRender(r IRender) {
	this.contextRender = r
}
func (this *AppLayout) SetFooterRender(r IRender) {
	this.contextRender = r
}
func (this *AppLayout) SetBottomRender(r IRender) {
	this.contextRender = r
}

func (this *AppLayout) SetHeadLayout(h *HeadLayout) {
	this.headLayout = h
}

func (this *AppLayout) RenderLayout(writer io.Writer) {
	writer.Write([]byte("<!DOCTYPE HTML>\n<html>\n"))
	this.headLayout.RenderLayout(writer)
	writer.Write([]byte("<body>\n"))
	this.topRender.Render(writer)
	this.headerRender.Render(writer)
	this.contextRender.Render(writer)
	this.footerRender.Render(writer)
	this.bottomRender.Render(writer)
	this.headLayout.RenderBottomJs(writer)

	writer.Write([]byte("\n</body>\n</html>"))
}

type HeadLayout struct {
	JsPosition     string
	Title          string
	HeadItemRender IRender
	CssRender      IRender
	JsRender       IRender
}

func (this *HeadLayout) RenderLayout(writer io.Writer) {
	writer.Write([]byte("<head>\n"))
	this.HeadItemRender.Render(writer)
	this.CssRender.Render(writer)
	if this.JsPosition == "head" {
		this.JsRender.Render(writer)
	}
	writer.Write([]byte("<title>" + this.Title + "</title>\n</head>\n"))
}

func (this *HeadLayout) RenderBottomJs(writer io.Writer) {
	if this.JsPosition != "head" {
		this.JsRender.Render(writer)
	}
}
