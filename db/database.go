package db

import (
	"bytes"
	"database/sql"
	"github.com/jiorry/gos/log"
	"reflect"
	"strconv"
	"time"
)

var bQuestionMark []byte = []byte("?")
var bEqual []byte = []byte("=")
var bDollar []byte = []byte("$")

type IDriver interface {
	ConnectString() string
	SetConnectString(string)
	QuoteField(string) string
	LastInsertId(string, string) string
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

func (this *Database) isError(err error) bool {
	if err != nil {
		log.App.Error(err)
		return true
	}
	return false
}

// Excute query on db prepare mode
func (this *Database) QueryPrepare(bSql []byte, args ...interface{}) (DataSet, error) {
	return this.QueryPrepareX(nil, bSql, args...)
}
func (this *Database) QueryPrepareX(cls interface{}, bSql []byte, args ...interface{}) (DataSet, error) {
	sqlstr := string(this.AdaptSql(bSql))
	dblog.Sql(sqlstr, args)
	s, err := this.Conn.Prepare(sqlstr)
	if this.isError(err) {
		return nil, err
	}
	rows, err := s.Query(args...)
	if this.isError(err) {
		return nil, err
	}
	defer rows.Close()

	dataset, err := scanRows(cls, rows)
	if this.isError(err) {
		return nil, err
	}
	return dataset, nil
}

func (this *Database) Query(bSql []byte, args ...interface{}) (DataSet, error) {
	return this.QueryX(nil, bSql, args...)
}

// Query from database, return DataSet result collection.
func (this *Database) QueryX(cls interface{}, bSql []byte, args ...interface{}) (DataSet, error) {
	sqlstr := string(this.AdaptSql(bSql))
	dblog.Sql(sqlstr, args)

	var rows *sql.Rows
	var err error
	rows, err = this.Conn.Query(sqlstr, args...)
	if this.isError(err) {
		return nil, err
	}
	defer rows.Close()

	dataset, err := scanRows(cls, rows)
	if this.isError(err) {
		return nil, err
	}
	return dataset, nil
}

// Excute sql command on db prepare mode
// In prepare mode, the sql command will be cached by database
func (this *Database) ExecPrepare(bSql []byte, args ...interface{}) (sql.Result, error) {
	sqlstr := string(this.AdaptSql(bSql))
	dblog.Sql(sqlstr, args)
	s, err := this.Conn.Prepare(sqlstr)
	if this.isError(err) {
		return nil, err
	}

	r, err := s.Exec(args...)
	if this.isError(err) {
		return nil, err
	}
	return r, nil
}

// Excute sql.
// If your has more than on sql command, it will only excute the first.
func (this *Database) Exec(bSql []byte, args ...interface{}) (sql.Result, error) {
	sqlstr := string(this.AdaptSql(bSql))
	dblog.Sql(sqlstr, args)
	r, err := this.Conn.Exec(sqlstr, args...)
	if this.isError(err) {
		return nil, err
	}
	return r, nil
}

func (this *Database) AdaptSql(bSql []byte) []byte {
	if this.DriverName == "postgres" {
		arr := bytes.Split(bSql, bQuestionMark)
		l := len(arr)
		s := bytes.Buffer{}
		s.Write(arr[0])

		for i := 1; i < l; i++ {
			s.Write(bDollar)
			s.WriteString(strconv.Itoa(i))
			s.Write(arr[i])
		}
		return s.Bytes()
	} else {
		return bSql
	}
}

func scanRows(cls interface{}, rows *sql.Rows) (DataSet, error) {
	var err error
	var dataset DataSet
	if cls == nil {
		dataset, err = ScanRowsToMap(rows)
	} else {
		var sm *structMaps
		switch inst := cls.(type) {
		case *structMaps:
			sm = inst
		default:
			sm = &structMaps{}
			sm.SetTarget(cls)
		}
		dataset, err = sm.ScanRowsToStruct(rows)
	}
	if err != nil {
		return nil, err
	}
	return dataset, nil
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

func (this DataSet) DataRow(index int) DataRow {
	if d, ok := this[index].(DataRow); ok {
		return d
	}
	return nil
}
func (this DataSet) Search(field string, value interface{}) interface{} {
	var datarow DataRow
	isDataRow := false
	isInit := false

	for _, r := range this {
		if !isInit {
			_, isDataRow = r.(DataRow)
			isInit = false
		}
		if isDataRow {
			datarow = r.(DataRow)
			if datarow[field] == value {
				return datarow
			}
		} else {
			val := reflect.ValueOf(r)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if v := val.FieldByName(field).Interface(); v == value {
				return v
			}
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
