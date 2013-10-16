package db

import (
	"bytes"
)

// Counter builder
type CounterBuilder struct {
	table string
	builder
}

func (this *CounterBuilder) Table(t string) *CounterBuilder {
	this.table = t
	return this
}
func (this *CounterBuilder) Query(cond string, args ...interface{}) (int64, error) {
	s := bytes.Buffer{}
	driver := this.builder.GetDatabase().Driver
	s.WriteString("select count(1) as count from ")
	s.WriteString(driver.QuoteField(this.table))
	if cond != "" {
		s.WriteString(" where ")
		s.WriteString(cond)
	}

	r, err := this.GetDatabase().QueryPrepare(s.Bytes(), args...)
	if err != nil {
		return -1, err
	}
	return r[0].(DataRow).GetInt64("count"), nil
}
