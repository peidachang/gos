package db

import (
	"bytes"
	"fmt"
	"github.com/jiorry/gos/cache"
	"github.com/jiorry/gos/util"
)

// Query builder
type QueryBuilder struct {
	table      string
	field      string
	where      *parpareParams
	order      string
	limit      int
	offset     int
	cache      bool
	expire     int
	ctype      string
	dataStruct interface{}

	builder
}

func (this *QueryBuilder) Table(t string) *QueryBuilder {
	this.table = t
	return this
}

func (this *QueryBuilder) Select(s string) *QueryBuilder {
	this.field = s
	return this
}

func (this *QueryBuilder) Where(s string, args ...interface{}) *QueryBuilder {
	this.where = &parpareParams{s, args}
	return this
}

func (this *QueryBuilder) Order(s string) *QueryBuilder {
	this.order = s
	return this
}

func (this *QueryBuilder) Page(page int, pageSize int) *QueryBuilder {
	this.offset = (page - 1) * pageSize
	this.limit = pageSize
	return this
}

func (this *QueryBuilder) Limit(n int) *QueryBuilder {
	this.limit = n
	return this
}

func (this *QueryBuilder) Offset(n int) *QueryBuilder {
	this.offset = n
	return this
}

func (this *QueryBuilder) DataStruct(v interface{}) *QueryBuilder {
	this.dataStruct = v
	return this
}

func (this *QueryBuilder) Cache(expire int) *QueryBuilder {
	this.cache = true
	this.expire = expire
	return this
}

func (this *QueryBuilder) cachekey() string {
	return util.MD5String(fmt.Sprintf("%v%v%v%v%v", this.field, this.where, this.limit, this.offset, this.order))
}

func (this *QueryBuilder) parse() string {
	sel := "*"
	conditions := ""
	order := ""
	limitoffset := ""

	s := bytes.Buffer{}

	if len(this.field) > 0 {
		sel = this.field
	}

	if this.where != nil {
		conditions = " where " + this.where.code
	}

	if this.limit > 0 || this.offset > 0 {
		limitoffset = this.GetDatabase().Driver.LimitOffsetStatement(this.limit, this.offset)
	}

	if len(this.order) > 0 {
		order = " order by " + this.order
	}

	driver := this.GetDatabase().Driver
	s.WriteString("select ")
	s.WriteString(sel)
	s.WriteString(" from ")
	s.WriteString(driver.QuoteField(this.table))
	s.WriteString(conditions)
	s.WriteString(order)
	s.WriteString(limitoffset)

	return s.String()
}

func (this *QueryBuilder) Query() (DataSet, error) {
	var key string
	if this.cache && cache.IsEnable() {
		key = this.cachekey()
		if exi, _ := cache.Exists(key); exi {
			return cacheGetDBResult(key)
		}
	}
	var r DataSet
	var err error
	if this.where == nil {
		if this.dataStruct == nil {
			r, err = this.GetDatabase().QueryPrepare(this.parse())
		} else {
			r, err = this.GetDatabase().QueryPrepareX(this.dataStruct, this.parse())
		}
	} else {
		if this.dataStruct == nil {
			r, err = this.GetDatabase().QueryPrepare(this.parse(), this.where.args...)
		} else {
			r, err = this.GetDatabase().QueryPrepareX(this.dataStruct, this.parse(), this.where.args...)
		}
	}

	if err != nil {
		return nil, err
	}
	if this.cache && cache.IsEnable() {
		cacheSet(key, r, this.expire)
	}
	return r, nil
}
func (this *QueryBuilder) First() (DataRow, error) {
	return this.QueryOne()
}
func (this *QueryBuilder) QueryOne() (DataRow, error) {
	this.limit = 1
	result, err := this.Query()
	if err != nil {
		return nil, err
	}

	var row DataRow
	if len(result) > 0 {
		row = result[0].(DataRow)
	} else {
		row = nil
	}
	return row, nil
}

func (this *QueryBuilder) Exists(s string, args ...interface{}) (bool, error) {
	this.where = &parpareParams{s, args}
	this.field = "1"
	r, err := this.Query()
	return len(r) > 0, err
}
