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

func (this *UpdateBuilder) Table(t string) *UpdateBuilder {
	this.table = t
	return this
}

func (this *UpdateBuilder) Where(cond string, args ...interface{}) *UpdateBuilder {
	this.where = &parpareParams{cond, args}
	return this
}

func (this *UpdateBuilder) Update(data interface{}) (sql.Result, error) {
	keys, values, _ := keyValueList(data)

	arr := make([][]byte, len(keys))
	str := []byte("=?")
	for i, _ := range keys {
		arr[i] = append(keys[i], str...)
	}
	var cond string = ""
	if this.where != nil {
		cond = " where " + this.where.code
		values = append(values, this.where.args...)
	}

	s := bytes.Buffer{}
	driver := this.builder.GetDatabase().Driver
	s.WriteString("update ")
	s.WriteString(driver.QuoteField(this.table))
	s.WriteString(" set ")
	s.Write(bytes.Join(arr, commaSplit))
	s.WriteString(cond)

	return this.GetDatabase().ExecPrepare(s.String(), values...)
}
