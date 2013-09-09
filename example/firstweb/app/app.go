package main

import (
	"./api"
	"./page"
	"./upload"
	"github.com/jiorry/gos/httpd"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	httpd.Init()

	httpd.AddRoute("/", (*page.IndexPage)(nil))

	// open api router
	httpd.AddWebServiceRoute("/web", (*websvr.ApiPublicService)(nil))

	// file upload router
	httpd.AddFileUploadRoute("/bbx", (*upload.UserFileUpload)(nil))

	httpd.Start()
}
