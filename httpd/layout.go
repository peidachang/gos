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

func (this *LayoutData) AddHeadItem(str string) {
	this.Head = append(this.Head, str)
}
func (this *LayoutData) AddJs(str string) {
	this.Head = append(this.Js, str)
}
func (this *LayoutData) AddCss(str string) {
	this.Head = append(this.Css, str)
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
	writer.Write([]byte("<!DOCTYPE HTML>\n"))
	writer.Write([]byte("<html>\n"))

	this.headLayout.RenderLayout(writer)

	writer.Write([]byte("<body>\n"))
	this.topRender.render(writer)
	this.headerRender.render(writer)
	this.contextRender.render(writer)
	this.footerRender.render(writer)
	this.bottomRender.render(writer)

	this.headLayout.RenderBottomJs(writer)
	writer.Write([]byte("</body>\n"))

	writer.Write([]byte("</html>"))
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
	this.HeadItemRender.render(writer)
	this.CssRender.render(writer)
	if this.JsPosition == "head" {
		this.JsRender.render(writer)
	}
	writer.Write([]byte("<title>" + this.Title + "</title>\n"))
	writer.Write([]byte("</head>\n"))
}

func (this *HeadLayout) RenderBottomJs(writer io.Writer) {
	if this.JsPosition != "head" {
		this.JsRender.render(writer)
	}
}
