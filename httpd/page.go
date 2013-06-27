package httpd

import (
	"github.com/jiorry/gos/log"
	"io"
	"os"
)

type Page struct {
	Ctx        *Context
	LayoutData *LayoutData
	Data       interface{}

	TopRender     IRender
	HeaderRender  IRender
	ContextRender IRender
	FooterRender  IRender
	BottomRender  IRender
}

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
}
func (this *Page) SetContext(ct *Context) {
	this.Ctx = ct
}

func (this *Page) RenderPage() {
	applayout := this.BuildLayout(this.Ctx.ResponseWriter)
	// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
	this.Ctx.ResponseWriter.WriteHeader(200)
	applayout.RenderLayout(this.Ctx.ResponseWriter)
}

func (this *Page) ToFile() {
	out, err := os.OpenFile("var/html/"+this.LayoutData.View+".html", os.O_TRUNC|os.O_CREATE, 0)
	if err != nil {
		log.App.Fatalln(err)
	}
	applayout := this.BuildLayout(out)
	applayout.RenderLayout(out)
}

func (this *Page) BuildLayout(writer io.Writer) *AppLayout {
	empty := &EmptyRender{}
	headLayout := &HeadLayout{Title: this.LayoutData.Title,
		HeadItemRender: empty, JsRender: empty, CssRender: empty}

	app := &AppLayout{HeadLayout: headLayout,
		TopRender: empty, HeaderRender: empty, ContextRender: empty,
		FooterRender: empty, BottomRender: empty, JsBottomRender: empty}

	if len(this.LayoutData.Head) > 0 {
		headLayout.HeadItemRender = &HeadItemRender{
			Data:   this.LayoutData.Head,
			Writer: writer}
	}

	if len(this.LayoutData.Css) > 0 {
		headLayout.CssRender = &CssRender{
			Data:      this.LayoutData.Css,
			Timestamp: this.LayoutData.Timestamp,
			Writer:    writer}
	}

	if len(this.LayoutData.Js) > 0 {
		if this.LayoutData.JsPosition == "head" {
			headLayout.JsRender = &JsRender{
				Data:      this.LayoutData.Js,
				Timestamp: this.LayoutData.Timestamp,
				Writer:    writer}
		} else {
			app.JsBottomRender = &JsRender{
				Data:      this.LayoutData.Js,
				Timestamp: this.LayoutData.Timestamp,
				Writer:    writer}
		}
	}

	if len(this.LayoutData.View) > 0 {
		app.ContextRender = &TemplateRender{
			View:   this.LayoutData.View,
			Data:   this.Data,
			Writer: writer}
	}

	return app
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
