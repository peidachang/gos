package db

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/jiorry/gos/cache"
	"github.com/jiorry/gos/log"
	"github.com/jiorry/gos/util"
	"strings"
	"time"
)

func init() {
	gob.Register(DataRow{})
	gob.Register(DataSet{})
	gob.Register(time.Time{})
}

type builderBase struct {
	database *Database
}

func (this *builderBase) GetDatabase() *Database {
	if this.database == nil {
		this.database = Current()
	}
	return this.database
}

func (this *builderBase) SetDatabase(d *Database) {
	this.database = d
}

type parpareParams struct {
	code string
	args []interface{}
}

// Query builder
type QueryBuilder struct {
	table  string
	field  string
	where  *parpareParams
	order  string
	limit  int
	offset int
	cache  bool
	expire int
	ctype  string

	builderBase
}

func (this *QueryBuilder) Table(t string) *QueryBuilder {
	this.table = t
	return this
}

func (this *QueryBuilder) First() *QueryBuilder {
	this.limit = 1
	this.offset = 0
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

func (this *QueryBuilder) Limit(n int) *QueryBuilder {
	this.limit = n
	return this
}

func (this *QueryBuilder) Offset(n int) *QueryBuilder {
	this.offset = n
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

	return "select " + sel + " from " + this.table + conditions + order + limitoffset
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
		r, err = this.GetDatabase().QueryPrepare(this.parse())
	} else {
		r, err = this.GetDatabase().QueryPrepare(this.parse(), this.where.args...)
	}

	if err != nil {
		return nil, err
	}
	if this.cache && cache.IsEnable() {
		cacheSet(key, r, this.expire)
	}
	return r, nil
}

func (this *QueryBuilder) QueryOne() (DataRow, error) {
	result, err := this.Query()
	if err != nil {
		return nil, err
	}

	var row DataRow
	if len(result) > 0 {
		row = result[0]
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

// insert builder
type InsertBuilder struct {
	table string
	builderBase
}

func (this *InsertBuilder) Table(t string) *InsertBuilder {
	this.table = t
	return this
}

func (this *InsertBuilder) parse(data DataRow) (code string, values []interface{}) {
	keys, values, stmts := keyValueList(data)
	code = "insert into " + this.table + " (" + strings.Join(keys, ",") + ") values (" + strings.Join(stmts, ",") + ")"
	return
}

func (this *InsertBuilder) Insert(row DataRow) (sql.Result, error) {
	code, args := this.parse(row)
	return this.GetDatabase().ExecPrepare(code, args...)
}

func (this *InsertBuilder) InsertM(rows DataSet) {
	for _, r := range rows {
		this.Insert(r)
	}
}

// Update builder
type UpdateBuilder struct {
	table string
	where *parpareParams
	builderBase
}

func (this *UpdateBuilder) Table(t string) *UpdateBuilder {
	this.table = t
	return this
}

func (this *UpdateBuilder) Where(cond string, args ...interface{}) *UpdateBuilder {
	this.where = &parpareParams{cond, args}
	return this
}

func (this *UpdateBuilder) Update(data DataRow) (sql.Result, error) {
	keys, values, _ := keyValueList(data)
	arr := make([]string, len(data))
	for i, _ := range keys {
		arr[i] = keys[i] + "=?"
	}
	var cond string = ""
	if this.where != nil {
		cond = " where " + this.where.code
		values = append(values, this.where.args...)
	}

	code := "update " + this.table + " set " + strings.Join(arr, ",") + cond
	return this.GetDatabase().ExecPrepare(code, values...)
}

// Delete builder
type DeleteBuilder struct {
	table string
	builderBase
}

func (this *DeleteBuilder) Table(t string) *DeleteBuilder {
	this.table = t
	return this
}

func (this *DeleteBuilder) Delete(cond string, args ...interface{}) (sql.Result, error) {
	if cond != "" {
		cond = " where " + cond
	}

	return this.GetDatabase().ExecPrepare("delete from "+this.table+cond, args...)
}

// Counter builder
type CounterBuilder struct {
	table string
	builderBase
}

func (this *CounterBuilder) Table(t string) *CounterBuilder {
	this.table = t
	return this
}
func (this *CounterBuilder) Query(cond string, args ...interface{}) (int64, error) {
	if cond != "" {
		cond = " where " + cond
	}
	r, err := this.GetDatabase().QueryPrepare("select count(1) as count from "+this.table+cond, args...)
	if err != nil {
		return 0, err
	}
	return r[0].GetInt64("count"), nil
}

func keyValueList(data DataRow) (keys []string, values []interface{}, stmts []string) {
	length := len(data)
	keys = make([]string, length)
	values = make([]interface{}, length)
	stmts = make([]string, length)
	i := 0
	for k, v := range data {
		keys[i] = k
		stmts[i] = "?"
		values[i] = v
		i++
	}

	return
}

func cacheSet(key string, value interface{}, expire int) error {
	if !cache.IsEnable() {
		return nil
	}

	if !cache.IsEnable() {
		return nil
	}
	v, err := gobEncode(value)
	if err != nil {
		log.App.Crit(err)
		return err
	}
	err = cache.Set(key, v, expire)
	if err != nil {
		log.App.Crit(err)
	}
	return err
}

func cacheGetDBResult(key string) (DataSet, error) {
	out := DataSet{}
	reply, err := cache.Get(key)
	if reply == nil || err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(reply.([]byte)))
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func cacheGetDBRow(key string) (DataRow, error) {
	var out = DataRow{}
	reply, err := cache.Get(key)
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(reply.([]byte)))
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func gobEncode(obj interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return []byte(""), err
	}
	return buf.Bytes(), nil
}
