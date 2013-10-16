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

	RenderFunc func(*AppLayout, io.Writer)
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

func (this *AppLayout) GetTopRender() IRender {
	return this.topRender
}
func (this *AppLayout) SetTopRender(r IRender) {
	if r == nil {
		this.topRender = RenderNothing
	} else {
		this.topRender = r
	}
}

func (this *AppLayout) GetHeaderRender() IRender {
	return this.headerRender
}
func (this *AppLayout) SetHeaderRender(r IRender) {
	if r == nil {
		this.headerRender = RenderNothing
	} else {
		this.headerRender = r
	}
}

func (this *AppLayout) GetContextRender() IRender {
	return this.contextRender
}
func (this *AppLayout) SetContextRender(r IRender) {
	if r == nil {
		this.contextRender = RenderNothing
	} else {
		this.contextRender = r
	}
}

func (this *AppLayout) GetFooterRender() IRender {
	return this.footerRender
}
func (this *AppLayout) SetFooterRender(r IRender) {
	if r == nil {
		this.footerRender = RenderNothing
	} else {
		this.footerRender = r
	}
}

func (this *AppLayout) GetBottomRender() IRender {
	return this.bottomRender
}
func (this *AppLayout) SetBottomRender(r IRender) {
	if r == nil {
		this.bottomRender = RenderNothing
	} else {
		this.bottomRender = r
	}
}

func (this *AppLayout) GetHeadLayout() *HeadLayout {
	return this.headLayout
}
func (this *AppLayout) SetHeadLayout(h *HeadLayout) {
	this.headLayout = h
}

func (this *AppLayout) RenderLayout(writer io.Writer) {
	if this.RenderFunc != nil {
		this.RenderFunc(this, writer)
		return
	}
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
