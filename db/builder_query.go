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
	dataStruct *structMaps

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

func (this *QueryBuilder) Struct(v interface{}) *QueryBuilder {
	ds := &structMaps{}
	ds.SetTarget(v)
	this.dataStruct = ds
	return this
}

func (this *QueryBuilder) Cache(expire int) *QueryBuilder {
	this.cache = true
	this.expire = expire
	return this
}

func (this *QueryBuilder) cachekey() []byte {
	s := bytes.Buffer{}
	s.WriteString(this.table)
	this.writeField(&s)
	s.WriteString(fmt.Sprintf("%d%d", this.limit, this.offset))
	s.WriteString(this.order)
	if this.where != nil {
		s.WriteString(fmt.Sprintf("%s%v", this.where.code, this.where.args))
	}
	return util.MD5(s.Bytes())
}

func (this *QueryBuilder) ClearCache() error {
	return cacheDel(this.cachekey())
}

func (this *QueryBuilder) parse() []byte {
	s := bytes.Buffer{}
	driver := this.GetDatabase().Driver
	s.WriteString("select ")

	this.writeField(&s)

	s.WriteString(" from ")
	s.WriteString(driver.QuoteField(this.table))

	if this.where != nil {
		s.WriteString(" where ")
		s.WriteString(this.where.code)
	}
	if this.order != "" {
		s.WriteString(" order by ")
		s.WriteString(this.order)
	}

	if this.limit > 0 && this.offset > 0 {
		s.WriteString(fmt.Sprintf(" limit %d offset %d", this.limit, this.offset))
	} else if this.limit > 0 {
		s.WriteString(fmt.Sprintf(" limit %d", this.limit))
	} else if this.offset > 0 {
		s.WriteString(fmt.Sprintf(" offset %d", this.offset))
	}

	return s.Bytes()
}

func (this *QueryBuilder) Query() (DataSet, error) {
	var key []byte
	if this.cache && cache.IsEnable() {
		key = this.cachekey()
		if exi, _ := cache.Exists(key); exi {
			return cacheGet(key, this.dataStruct)
		}
	}
	var r DataSet
	var err error
	sql := string(this.parse())
	if this.where == nil {
		if this.dataStruct == nil {
			r, err = this.GetDatabase().QueryPrepare(sql)
		} else {
			r, err = this.GetDatabase().QueryPrepareX(this.dataStruct, sql)
		}
	} else {
		if this.dataStruct == nil {
			r, err = this.GetDatabase().QueryPrepare(sql, this.where.args...)
		} else {
			r, err = this.GetDatabase().QueryPrepareX(this.dataStruct, sql, this.where.args...)
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

func (this *QueryBuilder) writeField(s *bytes.Buffer) {
	if this.dataStruct != nil {
		s.Write(bytes.Join(this.dataStruct.Fields(), commaSplit))
	} else if this.field != "" {
		s.WriteString(this.field)
	} else {
		s.WriteString("*")
	}
}
