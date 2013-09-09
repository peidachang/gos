package page

import (
	"github.com/jiorry/gos/httpd"
)

type IndexPage struct {
	httpd.Page
}

type indexdata struct {
	PageData string
}

func (this *IndexPage) Init() {
	this.Page.Init()
	this.Page.Title = "My First Go Web"
	this.SetView("index")
	this.AddCss("site.css")
	this.AddJs("jquery.js")

	this.Layout.HeaderView(httpd.SiteTheme, "_header", &renddata{"header"})
	this.Layout.FooterView(httpd.SiteTheme, "_footer", &renddata{"footer"})

	d := &indexdata{}
	d.PageData = "This value is from page struct"

	this.Data = d
}

type renddata struct {
	Name string
}
