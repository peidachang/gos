package db

import (
	"bytes"
	"database/sql"
)

// Delete builder
type DeleteBuilder struct {
	builder
	table string
	where *parpareParams
}

func (d *DeleteBuilder) Table(t string) *DeleteBuilder {
	d.table = t
	return d
}

func (d *DeleteBuilder) parse() ([]byte, []interface{}) {
	var args []interface{}

	s := bytes.Buffer{}
	driver := d.builder.GetDatabase().Driver
	s.WriteString("delete")
	s.WriteString(" from ")
	s.WriteString(driver.QuoteField(d.table))
	if d.where != nil {
		s.WriteString(" where ")
		s.WriteString(d.where.code)
		args = d.where.args
	}
	return s.Bytes(), args
}

func (d *DeleteBuilder) Where(cond string, args ...interface{}) *DeleteBuilder {
	d.where = &parpareParams{cond, args}
	return d
}

func (d *DeleteBuilder) Delete() (sql.Result, error) {
	s, args := d.parse()
	return d.GetDatabase().ExecPrepare(s, args...)
}

func (d *DeleteBuilder) Tx() *TxItem {
	sql, args := d.parse()
	return &TxItem{sql, args, nil}
}
