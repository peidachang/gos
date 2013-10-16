package db

import (
	"bytes"
	"database/sql"
)

// Update builder
type UpdateBuilder struct {
	table string
	where *parpareParams
	builder
}

func (u *UpdateBuilder) Table(t string) *UpdateBuilder {
	u.table = t
	return u
}

func (u *UpdateBuilder) parse(data interface{}) ([]byte, []interface{}) {
	keys, values, _ := keyValueList("update", data)

	s := bytes.Buffer{}
	driver := u.builder.GetDatabase().Driver
	s.WriteString("update ")
	s.WriteString(driver.QuoteField(u.table))
	s.WriteString(" set ")
	s.Write(bytes.Join(keys, commaSplit))
	if u.where != nil {
		s.WriteString(" where ")
		s.WriteString(u.where.code)
		values = append(values, u.where.args...)
	}
	return s.Bytes(), values
}

func (u *UpdateBuilder) Where(cond string, args ...interface{}) *UpdateBuilder {
	u.where = &parpareParams{cond, args}
	return u
}

func (u *UpdateBuilder) Update(data interface{}) (sql.Result, error) {
	sql, values := u.parse(data)
	return u.GetDatabase().ExecPrepare(sql, values...)
}

func (u *UpdateBuilder) Tx(data interface{}) *TxItem {
	sql, args := u.parse(data)
	return &TxItem{sql, args, nil}
}
