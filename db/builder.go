package db

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"github.com/jiorry/gos/cache"
	"github.com/jiorry/gos/log"
	"reflect"
	"time"
)

func init() {
	gob.Register(DataRow{})
	gob.Register(DataSet{})
	gob.Register(time.Time{})
}

type builder struct {
	database *Database
}

func (this *builder) GetDatabase() *Database {
	if this.database == nil {
		this.database = Current()
	}
	return this.database
}

func (this *builder) SetDatabase(d *Database) {
	this.database = d
}

type parpareParams struct {
	code string
	args []interface{}
}

func keyValueList(data interface{}) (keys [][]byte, values []interface{}, stmts [][]byte) {
	switch data.(type) {
	case DataRow, map[string]interface{}:
		inst := data.(DataRow)
		l := len(inst)
		keys = make([][]byte, l)
		values = make([]interface{}, l)
		stmts = make([][]byte, l)
		i := 0
		for k, v := range inst {
			keys[i] = []byte(k)
			stmts[i] = []byte("?")
			values[i] = v
			i++
		}
	default:
		sm := &structMaps{}
		sm.SetTarget(data)
		keys, values, stmts = sm.KeyValueList()
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

func rowsToMap(rows *sql.Rows) (DataSet, error) {
	cols, _ := rows.Columns()
	colsNum := len(cols)

	result := DataSet{}
	var err error
	var row, tem []interface{}
	var rowData map[string]interface{}

	for rows.Next() {
		row = make([]interface{}, colsNum)
		tem = make([]interface{}, colsNum)

		for i := range row {
			tem[i] = &row[i]
		}

		if err = rows.Scan(tem...); err != nil {
			log.App.Error(err)
			return nil, err
		}

		rowData = make(map[string]interface{})
		for i := range cols {
			rowData[cols[i]] = row[i]
		}

		result = append(result, rowData)
	}

	if err = rows.Err(); err != nil {
		log.App.Error(err)
		log.App.Stack()
		return nil, err
	}
	return result, nil
}

type structMaps struct {
	typ        reflect.Type
	val        reflect.Value
	fieldIndex map[string][]int
	fields     []string
}

func (this *structMaps) SetTarget(cls interface{}) {
	this.typ = reflect.TypeOf(cls)
	if this.typ.Kind() == reflect.Ptr {
		this.typ = this.typ.Elem()
	}

	this.val = reflect.ValueOf(cls)
	if this.val.Kind() == reflect.Ptr {
		this.val = this.val.Elem()
	}
}
func (this *structMaps) buildFieldInfo() {
	n := this.typ.NumField()
	this.fieldIndex = make(map[string][]int)
	this.fields = make([]string, n)
	for i := 0; i < n; i++ {
		f := this.typ.Field(i)
		field := f.Tag.Get("db")
		if field == "" {
			field = f.Name
		}
		this.fields[i] = field
		this.fieldIndex[field] = f.Index
	}
}
func (this *structMaps) Fields() []string {
	if this.fields == nil {
		this.buildFieldInfo()
	}
	return this.fields
}
func (this *structMaps) GetFieldIndex(colName string) []int {
	if this.fieldIndex == nil {
		this.buildFieldInfo()
	}
	for k, v := range this.fieldIndex {
		if k == colName {
			return v
		}
	}
	return make([]int, 0)
}

func (this *structMaps) ScanRowsToStruct(rows *sql.Rows) (DataSet, error) {
	cols, _ := rows.Columns()
	result := DataSet{}
	var err error
	var fieldIndex []int
	values := make([]interface{}, len(cols))

	for rows.Next() {
		rowStruct := reflect.New(this.typ)

		for i, c := range cols {
			fieldIndex = this.GetFieldIndex(c)
			values[i] = rowStruct.Elem().FieldByIndex(fieldIndex).Addr().Interface()
		}

		if err = rows.Scan(values...); err != nil {
			log.App.Error(err)
			return nil, err
		}

		result = append(result, rowStruct.Interface())
	}

	if err = rows.Err(); err != nil {
		log.App.Error(err)
		log.App.Stack()
		return nil, err
	}
	return result, nil
}

func (this *structMaps) KeyValueList() (keys [][]byte, values []interface{}, stmts [][]byte) {
	l := this.typ.NumField()
	keys = make([][]byte, l)
	values = make([]interface{}, l)
	stmts = make([][]byte, l)

	for i := 0; i < l; i++ {
		keys[i] = []byte(this.typ.Field(i).Name)
		values[i] = this.val.Field(i).Interface()
		stmts[i] = []byte("?")
	}
	return
}
