package httpd

import (
	"errors"
	"github.com/jiorry/gos/util"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

type Upload struct {
	parent         interface{}
	Ctx            *Context
	StorePath      string
	ExtAllowedList []string
}

func (this *Upload) Prepare(ct *Context, p interface{}) {
	this.Ctx = ct
	this.parent = p
}

func (this *Upload) Init() {
	this.StorePath = "static/uploads"
}

func (this *Upload) DoUpload() {

}

func (this *Upload) ParseFormFile(field string) (*OriginFile, error) {
	fn, header, err := this.Ctx.Request.FormFile(field)
	defer fn.Close()
	if err != nil {
		return nil, err
	}

	if len(this.Ctx.Request.Form["token"]) == 0 {
		return nil, errors.New("file token must be set!")
	}

	token := this.Ctx.Request.FormValue("token")
	println(token)
	return &OriginFile{FileName: header.Filename, File: fn, Token: token}, nil
}

func (this *Upload) ParseMultipartForm(field string) (*OriginFile, error) {
	f := this.Ctx.Request.MultipartForm.File[field]
	//v := this.Ctx.Request.MultipartForm.Value[this.NameField]

	if len(f) == 0 {
		return nil, errors.New("parameter is invalid!")
	}
	fileHeader := f[0]
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	return &OriginFile{FileName: fileHeader.Filename, File: file}, nil
}

func (this *Upload) Build(origin *OriginFile) (*StoreFile, error) {
	filename := origin.FileName
	var name string
	arr := strings.Split(filename, ".")
	ext := ""
	if len(arr) > 1 {
		ext = strings.ToLower("." + arr[len(arr)-1])
		name = filename[0 : len(filename)-len(ext)]
	} else {
		name = filename
	}

	if !util.InStringArray(this.ExtAllowedList, ext) {
		return nil, errors.New("file " + ext + " is forbidden")
	}

	return &StoreFile{
		StorePath:  strings.TrimPrefix(this.StorePath, "/"),
		StoreName:  util.Unique(),
		Ext:        ext,
		Name:       name,
		OriginFile: origin}, nil
}

type OriginFile struct {
	FileName string
	Token    string
	File     multipart.File
}

type StoreFile struct {
	Name       string
	Ext        string
	StorePath  string
	StoreName  string
	OriginFile *OriginFile
}

func (this *StoreFile) Store() error {
	if !strings.HasSuffix(this.StorePath, "/") {
		this.StorePath += "/"
	}

	file, err := os.Create(this.StorePath + this.StoreName + this.Ext)
	if err != nil {
		return err
	}
	defer file.Close()

	io.Copy(file, this.OriginFile.File)
	return nil
}

func (this *StoreFile) CreateFolderIfNotExists() {
	if _, err := os.Stat(this.StorePath); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(this.StorePath, os.ModeDir)
		}
	}
}
