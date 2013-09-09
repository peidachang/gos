package websvr

import (
	"./../model"
	"github.com/jiorry/gos/db"
	"github.com/jiorry/gos/httpd"
)

type ApiPublicService struct {
	httpd.WebService
}

func (this *ApiPublicService) FindByTitle(args httpd.MapData) (db.DataSet, error) {
	title := args.GetString("Title")
	q := db.QueryBuilder{}
	q.Table("Products").Struct(&model.ProductModel{}).Where("Title=?", title)
	d, _ := q.Query()
	return d, nil
}
