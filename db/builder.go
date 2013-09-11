package db

import (
	"database/sql"
	"encoding/json"
	"github.com/jiorry/gos/cache"
	"github.com/jiorry/gos/log"
	"reflect"
)

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

func ScanRowsToMap(rows *sql.Rows) (DataSet, error) {
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
	fields     [][]byte
}

func (this *structMaps) SetTarget(cls interface{}) {
	this.typ = reflect.TypeOf(cls)

	this.val = reflect.ValueOf(cls)
	if this.val.Kind() == reflect.Ptr {
		this.val = this.val.Elem()
	}
}
func (this *structMaps) GetTypeElem() reflect.Type {
	if this.typ.Kind() == reflect.Ptr {
		return this.typ.Elem()
	} else {
		return this.typ
	}
}

func (this *structMaps) GetType() reflect.Type {
	return this.typ
}

func (this *structMaps) GetValue() reflect.Value {
	return this.val
}

func (this *structMaps) buildFieldInfo() {
	typ := this.GetTypeElem()
	n := typ.NumField()
	this.fieldIndex = make(map[string][]int)
	this.fields = make([][]byte, n)
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		field := f.Tag.Get("db")
		if field == "" {
			field = f.Name
		}
		this.fields[i] = []byte(field)
		this.fieldIndex[field] = f.Index
	}
}
func (this *structMaps) Fields() [][]byte {
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
		rowStruct := reflect.New(this.GetTypeElem())

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
	typ := this.GetTypeElem()
	l := typ.NumField()
	keys = make([][]byte, l)
	values = make([]interface{}, l)
	stmts = make([][]byte, l)

	for i := 0; i < l; i++ {
		keys[i] = []byte(typ.Field(i).Name)
		values[i] = this.val.Field(i).Interface()
		stmts[i] = []byte("?")
	}
	return
}

func cacheDel(bkey []byte) error {
	return cache.Delete(bkey)
}

func cacheSet(bkey []byte, value interface{}, expire int) error {
	if !cache.IsEnable() {
		return nil
	}

	if !cache.IsEnable() {
		return nil
	}
	v, err := json.Marshal(value)
	if err != nil {
		log.App.Crit(err)
		return err
	}
	err = cache.Set(bkey, v, expire)
	if err != nil {
		log.App.Crit(err)
	}
	return err
}

func cacheGet(bkey []byte, smap *structMaps) (DataSet, error) {
	reply, err := cache.Get(bkey)
	if reply == nil || err != nil {
		return nil, err
	}

	if smap == nil {
		d := DataSet{}
		if err := json.Unmarshal(reply.([]byte), &d); err != nil {
			return nil, err
		}
		return d, nil
	} else {
		styp := reflect.SliceOf(smap.GetType())
		obj := reflect.New(styp).Interface()
		if err := json.Unmarshal(reply.([]byte), obj); err != nil {
			return nil, err
		}
		val := reflect.ValueOf(obj).Elem()
		l := val.Len()
		d := make([]interface{}, l)
		for i := 0; i < l; i++ {
			d[i] = val.Index(i).Interface()
		}
		return d, nil
	}

}
