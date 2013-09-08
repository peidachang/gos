package db

import (
	"bytes"
	"database/sql"
	"strings"
)

// insert builder
type InsertBuilder struct {
	table string
	builder
}

func (this *InsertBuilder) Table(t string) *InsertBuilder {
	this.table = t
	return this
}

func (this *InsertBuilder) parse(data DataRow) (code string, values []interface{}) {
	keys, values, stmts := keyValueList(data)
	s := bytes.Buffer{}
	driver := this.builder.GetDatabase().Driver
	s.WriteString("insert into ")
	s.WriteString(driver.QuoteField(this.table))
	s.WriteString(" (")
	s.WriteString(strings.Join(keys, ","))
	s.WriteString(") values (")
	s.WriteString(strings.Join(stmts, ","))
	s.WriteString(")")
	return s.String(), values
}

func (this *InsertBuilder) Insert(data interface{}) (sql.Result, error) {
	var row DataRow
	switch inst := data.(type) {
	case DataRow:
		row = inst
	case map[string]interface{}:
		row = DataRow(inst)
	default:
		row = structToDataRow(inst)
	}

	sql, args := this.parse(row)
	return this.GetDatabase().ExecPrepare(sql, args...)
}

func (this *InsertBuilder) InsertM(rows DataSet) {
	for _, r := range rows {
		this.Insert(r)
	}
}
