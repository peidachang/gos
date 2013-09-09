package upload

import (
	"fmt"
	"github.com/jiorry/gos/db"
	"github.com/jiorry/gos/httpd"
	"time"
)

type UserFileUpload struct {
	httpd.Upload
}

func (this *UserFileUpload) InitData() {
	this.Upload.InitData()
	// this.Upload.StorePath = "static/uploads"
	this.Upload.ExtAllowedList = []string{"", ".txt", ".zip", ".txt", ".7z", ".gz", ".tar", ".rar", ".jpg", ".jpeg", ".png", ".gif"}
}

func (this *UserFileUpload) DoUpload() {
	f, err := this.Upload.ParseFormFile("files[]")
	// // f, err := this.Upload.ParseMultipartForm("Filedata")
	if err != nil {
		println(err.Error())
		return
	}

	fn, err1 := this.Upload.Build(f)
	if err1 != nil {
		println(err1.Error())
		return
	}
	table := "ContentFiles"
	folder := "c"
	now := time.Now()
	ts := now.Format("200601")

	fn.CreateFolderIfNotExists()
	fn.StorePath += "/" + folder
	fn.CreateFolderIfNotExists()
	fn.StorePath += "/" + ts
	fn.CreateFolderIfNotExists()

	fn.Store()

	row := db.DataRow{
		"Name":      fn.Name,
		"Ext":       fn.Ext,
		"FilePath":  fn.StorePath,
		"StoreName": fn.StoreName,
		"UserId":    1,
		"CreatedAt": now,
	}

	insert := &db.InsertBuilder{}
	result, err := insert.Table(table).Insert(row)
	fileId, err := result.LastInsertId()

	websvr := &httpd.WebService{Ctx: this.Ctx}

	rp := map[string]string{
		"fid":      fmt.Sprint(fileId),
		"filename": fn.Name + fn.Ext,
		"url":      httpd.StaticUrl + "/uploads/" + folder + "/" + ts + "/" + fn.StoreName + fn.Ext,
	}

	websvr.Reply(rp, err)
}
