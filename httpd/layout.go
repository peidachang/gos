package httpd

import (
	"io"
)

type LayoutData struct {
	View       string
	Title      string
	Head       []string
	Js         []string
	Css        []string
	Timestamp  string
	JsPosition string // head or end
}

func (this *LayoutData) AddHeadItem(items ...string) {
	this.Head = append(this.Head, items...)
}
func (this *LayoutData) AddJs(items ...string) {
	this.Js = append(this.Js, items...)
}
func (this *LayoutData) AddCss(items ...string) {
	this.Css = append(this.Css, items...)
}

type AppLayout struct {
	headLayout    *HeadLayout
	topRender     IRender
	headerRender  IRender
	contextRender IRender
	footerRender  IRender
	bottomRender  IRender
}

func (this *AppLayout) TopView(view string, data interface{}) {
	this.topRender = &TemplateRender{
		View: view,
		Data: data}
}
func (this *AppLayout) HeaderView(view string, data interface{}) {
	this.headerRender = &TemplateRender{
		View: view,
		Data: data}
}
func (this *AppLayout) FooterView(view string, data interface{}) {
	this.footerRender = &TemplateRender{
		View: view,
		Data: data}
}
func (this *AppLayout) BottomView(view string, data interface{}) {
	this.bottomRender = &TemplateRender{
		View: view,
		Data: data}
}
func (this *AppLayout) SetTop(r IRender) {
	this.contextRender = r
}
func (this *AppLayout) SetHeader(r IRender) {
	this.contextRender = r
}
func (this *AppLayout) SetContext(r IRender) {
	this.contextRender = r
}
func (this *AppLayout) SetFooter(r IRender) {
	this.contextRender = r
}
func (this *AppLayout) SetBottom(r IRender) {
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
