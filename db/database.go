package db

import (
	"database/sql"
	"github.com/jiorry/gos/log"
	"reflect"
	"strings"
	"time"
)

type IDriver interface {
	ConnectString() string
	LimitOffsetStatement(int, int) string
	QuoteField(string) string
}

type Database struct {
	Name       string
	Driver     IDriver
	DriverName string
	Conn       *sql.DB
}

// Connect database
func (this *Database) Connect() error {
	c, err := sql.Open(this.DriverName, this.Driver.ConnectString())
	if err != nil {
		log.App.Emerg(err)
		return err
	}
	this.Conn = c
	return nil
}

// Excute query on db prepare mode
func (this *Database) QueryPrepare(sqlstr string, args ...interface{}) (DataSet, error) {
	return this.QueryPrepareX(nil, sqlstr, args...)
}
func (this *Database) QueryPrepareX(cls interface{}, sqlstr string, args ...interface{}) (DataSet, error) {
	s, err := this.Conn.Prepare(sqlstr)
	if err != nil {
		log.App.Alert(err)
		log.App.Alert(sqlstr)
		log.App.Alert(args...)
		log.App.Stack()
		return nil, err
	}
	rows, err := s.Query(args...)
	if err != nil {
		log.App.Alert(err)
		log.App.Alert(sqlstr)
		log.App.Alert(args...)
		log.App.Stack()
		return nil, err
	}
	defer rows.Close()

	var result DataSet
	if cls == nil {
		result, err = rowToMap(rows)
	} else {
		result, err = rowToStruct(rows, cls)
	}
	if err != nil {
		return nil, err
	}
	dblog.Sql(sqlstr, args)

	return result, nil
}

func (this *Database) Query(sqlstr string, args ...interface{}) (DataSet, error) {
	return this.QueryX(nil, sqlstr, args...)
}

// Query from database, return DataSet result collection.
func (this *Database) QueryX(cls interface{}, sqlstr string, args ...interface{}) (DataSet, error) {
	if len(strings.TrimSpace(sqlstr)) == 0 {
		return nil, nil
	}

	var rows *sql.Rows
	var err error
	rows, err = this.Conn.Query(sqlstr, args...)

	if err != nil {
		log.App.Alert(err)
		log.App.Alert(sqlstr)
		log.App.Alert(args...)
		log.App.Stack()
		return nil, err
	}
	defer rows.Close()

	var result DataSet
	if cls == nil {
		result, err = rowToMap(rows)
	} else {
		result, err = rowToStruct(rows, cls)
	}
	if err != nil {
		return nil, err
	}

	dblog.Sql(sqlstr, args)

	return result, nil
}

// Excute sql command on db prepare mode
// In prepare mode, the sql command will be cached by database
func (this *Database) ExecPrepare(sqlstr string, args ...interface{}) (sql.Result, error) {
	s, _ := this.Conn.Prepare(sqlstr)
	r, err := s.Exec(args...)
	if err != nil {
		log.App.Alert("db exec error:", err, "\n", "sql:"+sqlstr+"|")
		log.App.Stack()
		return nil, err
	}
	return r, nil
}

// Excute sql.
// If your has more than on sql command, it will only excute the first.
func (this *Database) Exec(sqlstr string, args ...interface{}) (sql.Result, error) {
	r, err := this.Conn.Exec(sqlstr, args...)
	if err != nil {
		log.App.Alert("db exec error:", err, "\n", "sql:"+sqlstr+"|")
		log.App.Stack()
		return nil, err
	}
	return r, nil
}

type DataSet []interface{}
type DataRow map[string]interface{}

func (this DataRow) IsSet(key string) bool {
	_, ok := this[key]
	return ok
}
func (this DataRow) GetString(field string) string {
	if this[field] == nil {
		return ""
	}
	return this[field].(string)
}

func (this DataRow) GetInt64(field string) int64 {
	return this[field].(int64)
}

func (this DataRow) GetInt(field string) int {
	return this[field].(int)
}

func (this DataRow) GetFloat(field string) float64 {
	return this[field].(float64)
}

func (this DataRow) GetTime(field string) time.Time {
	if this[field] == nil {
		return time.Unix(0, 0)
	}
	return this[field].(time.Time)
}

func (this DataSet) Search(field string, value interface{}) DataRow {
	var row DataRow
	for _, r := range this {
		row = r.(DataRow)
		if row[field] == value {
			return row
		}
	}
	return nil
}

func (this DataSet) Encode() [][]interface{} {
	var value []interface{}
	var colSize int

	values := make([][]interface{}, len(this)+1)
	columns := make([]interface{}, 0)
	isFirst := true
	var row DataRow
	for i, r := range this {
		row = r.(DataRow)
		if isFirst {
			for k, _ := range row {
				columns = append(columns, k)
			}
			colSize = len(columns)
			values[i] = columns
			isFirst = false
		}

		value = make([]interface{}, colSize)
		n := 0
		for _, v := range row {
			value[n] = v
			n++
		}

		values[i+1] = value
	}

	return values
}

func rowToMap(rows *sql.Rows) (DataSet, error) {
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

type structFiledMaps struct {
	fieldInfos map[string][]int
}

func (this *structFiledMaps) Parse(typ reflect.Type) {
	this.fieldInfos = make(map[string][]int)
	n := typ.NumField()
	for i := 0; i < n; i++ {
		f := typ.Field(i)
		field := f.Tag.Get("db")
		if field == "" {
			field = f.Name
		}

		this.fieldInfos[field] = f.Index
	}
}
func (this *structFiledMaps) GetFieldIndex(colName string) []int {
	for k, v := range this.fieldInfos {
		if k == colName {
			return v
		}
	}
	return make([]int, 0)
}

func rowToStruct(rows *sql.Rows, cls interface{}) (DataSet, error) {
	typ := reflect.TypeOf(cls)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	sf := &structFiledMaps{}
	sf.Parse(typ)

	cols, _ := rows.Columns()
	result := DataSet{}
	var err error
	var fieldIndex []int
	values := make([]interface{}, len(cols))

	for rows.Next() {
		rowStruct := reflect.New(typ)

		for i, c := range cols {
			fieldIndex = sf.GetFieldIndex(c)
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

func structToDataRow(u interface{}) DataRow {
	typ := reflect.TypeOf(u)
	v := reflect.ValueOf(u)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		typ = typ.Elem()
	}
	datarow := DataRow{}
	for i := 0; i < v.NumField(); i++ {
		datarow[typ.Field(i).Name] = v.Field(i).Interface()
	}
	return datarow
}
