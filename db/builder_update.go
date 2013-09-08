package db

import (
	"bytes"
	"database/sql"
	"strings"
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
	var row DataRow
	switch inst := data.(type) {
	case DataRow:
		row = inst
	case map[string]interface{}:
		row = DataRow(inst)
	default:
		row = structToDataRow(inst)
	}

	keys, values, _ := keyValueList(row)

	arr := make([]string, len(row))
	for i, _ := range keys {
		arr[i] = keys[i] + "=?"
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
	s.WriteString(strings.Join(arr, ","))
	s.WriteString(cond)

	return this.GetDatabase().ExecPrepare(s.String(), values...)
}
