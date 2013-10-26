package httpd

import (
	"bytes"
	"github.com/jiorry/libs/log"
	"github.com/jiorry/libs/util"
	"net/http"
	"os"
	"reflect"
)

type PageCache struct {
	Type   string //none, file, cache
	Expire int64
}
type Page struct {
	View  *ThemeItem
	Title string
	Head  []string
	Js    []*ThemeItem
	Css   []*ThemeItem

	Timestamp   string
	JsPosition  string // head or end
	RequireAuth bool

	Cache  *PageCache
	Ctx    *Context
	Data   interface{}
	Layout *AppLayout
	parent interface{}
	auth   *UserAuth
}

type IPage interface {
	ToStaticFile()
}

func (p *Page) SetUserAuth(u *UserAuth) {
	p.auth = u
}

func (p *Page) GetUserAuth() *UserAuth {
	if p.auth == nil {
		p.auth = (&UserAuth{}).SetContext(p.Ctx)
	}
	return p.auth
}

func (p *Page) SetData(d interface{}) {
	p.Data = d
}

func (p *Page) SetView(viewName string) *Page {
	p.View = &ThemeItem{"", "template", viewName, nil}
	return p
}

func (p *Page) SetThemeView(theme string, viewName string) *Page {
	p.View = &ThemeItem{theme, "template", viewName, nil}
	return p
}

func (p *Page) AddHead(items ...string) *Page {
	p.Head = append(p.Head, items...)
	return p
}

func (p *Page) AddJsThemeItem(theme, value string, data map[string]string) *Page {
	p.Js = append(p.Js, &ThemeItem{theme, "js", value, data})
	return p
}

func (p *Page) AddThemeJs(theme string, items ...string) *Page {
	arr := make([]*ThemeItem, len(items))
	for i, _ := range items {
		arr[i] = &ThemeItem{theme, "js", items[i], nil}
	}
	p.Js = append(p.Js, arr...)
	return p
}
func (p *Page) AddJs(items ...string) *Page {
	return p.AddThemeJs("", items...)
}

func (p *Page) AddThemeCss(theme string, items ...string) *Page {
	arr := make([]*ThemeItem, len(items))
	for i, _ := range items {
		arr[i] = &ThemeItem{theme, "css", items[i], nil}
	}
	p.Css = append(p.Css, arr...)
	return p
}

func (p *Page) AddCss(items ...string) *Page {
	return p.AddThemeCss("", items...)
}

func (p *Page) Prepare(ct *Context, parent interface{}) {
	p.Ctx = ct
	p.parent = parent

	p.Cache = &PageCache{"none", 0}
	p.Js = make([]*ThemeItem, 0)
	p.Css = make([]*ThemeItem, 0)
	p.JsPosition = "head"
	p.Head = []string{`<meta charset="utf-8">`}

	// p.Data = make(map[string]interface{})

	p.Layout = &AppLayout{
		topRender:     RenderNothing,
		headerRender:  RenderNothing,
		contextRender: RenderNothing,
		footerRender:  RenderNothing,
		bottomRender:  RenderNothing}
}

func (p *Page) RenderPage() {
	// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
	// p.Ctx.ResponseWriter.WriteHeader(200)
	p.BuildLayout().RenderLayout(p.Ctx.ResponseWriter)
}

func (p *Page) CheckCache() int {
	if RunMode != "pro" {
		return CACHE_DISABLED
	}

	filename := "var/cache" + p.Ctx.Request.RequestURI + ".html"
	switch p.Cache.Type {
	case "file":
		if _, err := os.Stat(filename); err != nil {
			if os.IsNotExist(err) {
				return CACHE_NOT_FOUND
			}
			return CACHE_NOT_FOUND
		}
		http.ServeFile(p.Ctx.ResponseWriter, p.Ctx.Request, filename)
		return CACHE_FOUND
	}
	return CACHE_DISABLED
}

var cachePathList []string = make([]string, 0)

// check cache file, if cache file is exists, it will return content by cache and
// if cache file is not exists, it will create cache file
// filename=/content/cid?abc=123
// the dir /var/content and file /var/content/cid?abc=123 will be created.
func (p *Page) CachePage() {
	if p.Cache.Type == "file" {
		filename := "var/cache" + p.Ctx.Request.RequestURI + ".html"
		uri := []byte(p.Ctx.Request.RequestURI)
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

		p.savePageToFile(filename)
		http.ServeFile(p.Ctx.ResponseWriter, p.Ctx.Request, filename)
	}
}

func (p *Page) savePageToFile(filename string) {
	out, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE, 0)
	if err != nil {
		log.App.Fatalln(err)
	}
	p.BuildLayout().RenderLayout(out)
}
func (p *Page) ToStaticFile() {
	p.savePageToFile(httpServer.WebRoot + "/" + p.View.GetPath() + ".html")
}

func (p *Page) BuildLayout() *AppLayout {
	headLayout := &HeadLayout{
		JsPosition:     p.JsPosition,
		Title:          p.Title,
		HeadItemRender: RenderNothing,
		JsRender:       RenderNothing,
		CssRender:      RenderNothing}

	if len(p.Head) > 0 {
		headLayout.HeadItemRender = &HeadItemRender{
			Data: p.Head}
	}

	if len(p.Css) > 0 {
		headLayout.CssRender = &CssRender{
			Data: p.Css}
	}

	if len(p.Js) > 0 {
		headLayout.JsRender = &JsRender{
			Data: p.Js}
	}
	p.Layout.SetHeadLayout(headLayout)

	if p.View != nil {
		p.Layout.SetContextRender(&TemplateRender{
			View: p.View,
			Data: p.Data})
	}

	return p.Layout
}

func (p *Page) Auth()   {}
func (p *Page) Init()   {}
func (p *Page) Get()    {}
func (p *Page) Post()   {}
func (p *Page) Action() {}

type ThemeItem struct {
	Theme, Folder, Value string
	Data                 map[string]string
}

func (th *ThemeItem) GetPath() string {
	if th.Theme == "" {
		return "/" + th.Folder + "/" + th.Value
	}
	return "/themes/" + th.Theme + "/" + th.Folder + "/" + th.Value
}
func (th *ThemeItem) GetAssetsPath() string {
	if th.Theme == "" {
		return "/" + AssetsName + "/" + th.Folder + "/" + th.Value
	}
	return "/themes/" + th.Theme + "/" + AssetsName + "/" + th.Folder + "/" + th.Value
}

func StaticFiles(pages []IPage) {
	defer func() {
		if err := recover(); err != nil {
			log.App.Emerg(err)
		}
	}()

	for _, p := range pages {
		v := reflect.ValueOf(p)
		v.MethodByName("Prepare").Call([]reflect.Value{reflect.ValueOf(nil), v})
		p.ToStaticFile()
	}
}
