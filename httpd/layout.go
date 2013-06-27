package httpd

import (
	"io"
)

type AppLayout struct {
	HeadLayout     *HeadLayout
	TopRender      IRender
	HeaderRender   IRender
	ContextRender  IRender
	FooterRender   IRender
	BottomRender   IRender
	JsBottomRender IRender
}

func (this *AppLayout) RenderLayout(writer io.Writer) {
	writer.Write([]byte("<!DOCTYPE HTML>\n"))
	writer.Write([]byte("<html>\n"))

	this.HeadLayout.RenderLayout(writer)

	writer.Write([]byte("<body>\n"))
	this.TopRender.render()
	this.HeaderRender.render()
	this.ContextRender.render()
	this.FooterRender.render()
	this.BottomRender.render()
	this.JsBottomRender.render()
	writer.Write([]byte("</body>\n"))

	writer.Write([]byte("</html>"))
}

type HeadLayout struct {
	Title          string
	HeadItemRender IRender
	CssRender      IRender
	JsRender       IRender
}

func (this *HeadLayout) RenderLayout(writer io.Writer) {
	writer.Write([]byte("<head>\n"))
	this.HeadItemRender.render()
	this.CssRender.render()
	this.JsRender.render()
	writer.Write([]byte("<title>" + this.Title + "</title>\n"))
	writer.Write([]byte("</head>\n"))
}
