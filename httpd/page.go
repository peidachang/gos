package httpd

import (
	"github.com/jiorry/gos/log"
	"os"
)

var empty = &EmptyRender{}

type Page struct {
	Ctx        *Context
	LayoutData *LayoutData
	Data       interface{}
	Layout     *AppLayout
}

type IPage interface {
	Init()
	ToFile()
}

func DefaultData() *LayoutData {
	p := &LayoutData{JsPosition: "head"}
	p.Head = []string{
		`<meta charset="utf-8">`,
		`<script type="text/javascript">function addLoadEvent(func){var oldonload=window.myonload;if(func && typeof oldonload!="function")window.myonload=func;else{window.myonload=function(){if(oldonload) oldonload();if(func) func()}}}</script>`}
	return p
}

func CachePageToFile(pages []IPage) {
	for _, p := range pages {
		p.Init()
		p.ToFile()
	}
}

func (this *Page) Init() {
	this.LayoutData = DefaultData()
	this.Data = make(map[string]interface{})

	this.Layout = &AppLayout{
		topRender:     empty,
		headerRender:  empty,
		contextRender: empty,
		footerRender:  empty,
		bottomRender:  empty}
}
func (this *Page) SetContext(ct *Context) {
	this.Ctx = ct
}

func (this *Page) RenderPage() {
	// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
	// this.Ctx.ResponseWriter.WriteHeader(200)
	this.BuildLayout().RenderLayout(this.Ctx.ResponseWriter)
}

func (this *Page) ToFile() {
	out, err := os.OpenFile("var/html/"+this.LayoutData.View+".html", os.O_TRUNC|os.O_CREATE, 0)
	if err != nil {
		log.App.Fatalln(err)
	}
	this.BuildLayout().RenderLayout(out)
}

func (this *Page) BuildLayout() *AppLayout {
	headLayout := &HeadLayout{
		JsPosition:     this.LayoutData.JsPosition,
		Title:          this.LayoutData.Title,
		HeadItemRender: empty,
		JsRender:       empty,
		CssRender:      empty}

	if len(this.LayoutData.Head) > 0 {
		headLayout.HeadItemRender = &HeadItemRender{
			Data: this.LayoutData.Head}
	}

	if len(this.LayoutData.Css) > 0 {
		headLayout.CssRender = &CssRender{
			Data:      this.LayoutData.Css,
			Timestamp: this.LayoutData.Timestamp}
	}

	if len(this.LayoutData.Js) > 0 {
		headLayout.JsRender = &JsRender{
			Data:      this.LayoutData.Js,
			Timestamp: this.LayoutData.Timestamp}
	}
	this.Layout.SetHeadLayout(headLayout)

	if len(this.LayoutData.View) > 0 {
		this.Layout.SetContext(&TemplateRender{
			View: this.LayoutData.View,
			Data: this.Data})
	}

	return this.Layout
}

func (this *Page) Before() {

}
func (this *Page) Get() {

}
func (this *Page) Post() {

}
func (this *Page) After() {

}

type ThemeData struct {
	Css      string
	Js       string
	Template string
}

func (this *ThemeData) GetCss() string {
	if this.Css == "" {
		return "/css/"
	}
	return "/themes/" + this.Css + "/css/"
}
func (this *ThemeData) GetJs() string {
	if this.Js == "" {
		return "/js/"
	}
	return "/themes/" + this.Js + "/js/"
}
func (this *ThemeData) GetTemplate() string {
	if this.Js == "" {
		return "/"
	}
	return "/themes/" + this.Template + "/"
}
