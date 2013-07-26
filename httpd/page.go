package httpd

import (
	"bytes"
	"github.com/jiorry/gos/log"
	"github.com/jiorry/gos/util"
	"net/http"
	"os"
	"reflect"
)

var RenderNothing = &EmptyRender{}

const (
	CACHE_NOT_FOUND int = 0
	CACHE_FOUND     int = 1
	CACHE_DISABLED  int = -1
)

type PageCache struct {
	Type   string //none, file, cache
	Expire int64
}
type Page struct {
	RequireAuth bool
	Cache       *PageCache
	Ctx         *Context
	LayoutData  *LayoutData
	Data        interface{}
	Layout      *AppLayout
}

type IPage interface {
	ToStaticFile()
}

func DefaultData() *LayoutData {
	p := &LayoutData{JsPosition: "head"}
	p.Head = []string{
		`<meta charset="utf-8">`,
		`<script type="text/javascript">function addLoadFunction(func){var oldonload=window.onload;if(func && typeof oldonload!="function")window.onload=func;else{window.onload=function(){if(oldonload) oldonload();if(func) func()}}}</script>`}
	return p
}

func StaticFiles(pages []IPage) {
	defer func() {
		if err := recover(); err != nil {
			log.App.Err(err)
		}
	}()

	for _, p := range pages {
		v := reflect.ValueOf(p)
		v.MethodByName("Init").Call(nil)
		p.ToStaticFile()
	}
}

func (this *Page) Init() {
	this.Cache = &PageCache{"none", 0}
	this.LayoutData = DefaultData()
	this.Data = make(map[string]interface{})

	this.Layout = &AppLayout{
		topRender:     RenderNothing,
		headerRender:  RenderNothing,
		contextRender: RenderNothing,
		footerRender:  RenderNothing,
		bottomRender:  RenderNothing}
}
func (this *Page) SetContext(ct *Context) {
	this.Ctx = ct
}

func (this *Page) RenderPage() {
	// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
	// this.Ctx.ResponseWriter.WriteHeader(200)
	this.BuildLayout().RenderLayout(this.Ctx.ResponseWriter)
}

func (this *Page) CheckPageCache() int {
	if RunMode != "pro" {
		return CACHE_DISABLED
	}

	filename := "var/cache" + this.Ctx.Request.RequestURI + ".html"
	switch this.Cache.Type {
	case "file":
		if _, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				return CACHE_NOT_FOUND
			}
			return CACHE_NOT_FOUND
		}
		http.ServeFile(this.Ctx.ResponseWriter, this.Ctx.Request, filename)
		return CACHE_FOUND
	}
	return CACHE_DISABLED
}

var cachePathList []string = make([]string, 0)

// check cache file, if cache file is exists, it will return content by cache and
// if cache file is not exists, it will create cache file
// filename=/content/cid?abc=123
// the dir /var/content and file /var/content/cid?abc=123 will be created.
func (this *Page) CachePage() {
	if this.Cache.Type == "file" {
		filename := "var/cache" + this.Ctx.Request.RequestURI + ".html"
		uri := []byte(this.Ctx.Request.RequestURI)
		n := bytes.Index(uri, []byte("?"))
		var t []byte

		if n != -1 {
			t = bytes.Trim(uri[:n], "/")
		} else {
			t = bytes.Trim(uri, "/")
		}
		arr := bytes.Split(t, []byte("/"))
		// prepare dir, if dir is not exists , it will be create.
		path := "var/cache/"
		count := len(arr) - 1
		for i := 0; i < count; i++ {
			path += string(arr[i]) + "/"

			if util.InStringArray(cachePathList, path) {
				continue
			} else {
				cachePathList = append(cachePathList, path)
			}

			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					os.Mkdir(path, os.ModeDir)
				}
			}

		}

		this.savePageToFile(filename)
		http.ServeFile(this.Ctx.ResponseWriter, this.Ctx.Request, filename)
	}
}

func (this *Page) savePageToFile(filename string) {
	out, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE, 0)
	if err != nil {
		log.App.Fatalln(err)
	}
	this.BuildLayout().RenderLayout(out)
}
func (this *Page) ToStaticFile() {
	this.savePageToFile(httpServer.StaticDir + "/" + this.LayoutData.View + ".html")
}

func (this *Page) BuildLayout() *AppLayout {
	headLayout := &HeadLayout{
		JsPosition:     this.LayoutData.JsPosition,
		Title:          this.LayoutData.Title,
		HeadItemRender: RenderNothing,
		JsRender:       RenderNothing,
		CssRender:      RenderNothing}

	if len(this.LayoutData.Head) > 0 {
		headLayout.HeadItemRender = &HeadItemRender{
			Data: this.LayoutData.Head}
	}

	if len(this.LayoutData.Css) > 0 {
		headLayout.CssRender = &CssRender{
			Data: this.LayoutData.Css}
	}

	if len(this.LayoutData.Js) > 0 {
		headLayout.JsRender = &JsRender{
			Data: this.LayoutData.Js}
	}
	this.Layout.SetHeadLayout(headLayout)

	if len(this.LayoutData.View) > 0 {
		this.Layout.SetContext(&TemplateRender{
			View: this.LayoutData.View,
			Data: this.Data})
	}

	return this.Layout
}

func (this *Page) Auth()  {}
func (this *Page) Get()   {}
func (this *Page) Post()  {}
func (this *Page) After() {}

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
