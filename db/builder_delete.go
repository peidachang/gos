package db

import (
	"bytes"
	"database/sql"
)

// Delete builder
type DeleteBuilder struct {
	table string
	builder
	where *parpareParams
}

func (this *DeleteBuilder) Table(t string) *DeleteBuilder {
	this.table = t
	return this
}

func (this *DeleteBuilder) Where(cond string, args ...interface{}) *DeleteBuilder {
	this.where = &parpareParams{cond, args}
	return this
}

func (this *DeleteBuilder) Delete() (sql.Result, error) {
	var args []interface{}
	var cond string = ""
	if this.where != nil {
		cond = " where " + this.where.code
		args = this.where.args
	}

	s := bytes.Buffer{}
	driver := this.builder.GetDatabase().Driver
	s.WriteString("delete")
	s.WriteString(" from ")
	s.WriteString(driver.QuoteField(this.table))
	s.WriteString(cond)
	return this.GetDatabase().ExecPrepare(s.String(), args...)
}
