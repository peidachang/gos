package db

import (
	"bytes"
	"database/sql"
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

func (this *InsertBuilder) parse(data interface{}) (code string, values []interface{}) {
	keys, values, stmts := keyValueList(data)
	s := bytes.Buffer{}
	driver := this.builder.GetDatabase().Driver
	s.WriteString("insert into ")
	s.WriteString(driver.QuoteField(this.table))
	s.WriteString(" (")
	s.Write(bytes.Join(keys, commaSplit))
	s.WriteString(") values (")
	s.Write(bytes.Join(stmts, commaSplit))
	s.WriteString(")")
	return s.String(), values
}

func (this *InsertBuilder) Insert(data interface{}) (sql.Result, error) {
	sql, args := this.parse(data)
	return this.GetDatabase().ExecPrepare(sql, args...)
}

func (this *InsertBuilder) InsertM(rows DataSet) {
	for _, r := range rows {
		this.Insert(r)
	}
}
