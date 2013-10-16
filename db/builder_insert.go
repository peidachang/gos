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

func (ins *InsertBuilder) Table(t string) *InsertBuilder {
	ins.table = t
	return ins
}

func (ins *InsertBuilder) parse(data interface{}) (code []byte, values []interface{}) {
	keys, values, stmts := keyValueList("insert", data)
	s := bytes.Buffer{}
	driver := ins.builder.GetDatabase().Driver
	s.WriteString("insert into ")
	s.WriteString(driver.QuoteField(ins.table))
	s.WriteString(" (")
	s.Write(bytes.Join(keys, commaSplit))
	s.WriteString(") values (")
	s.Write(bytes.Join(stmts, commaSplit))
	s.WriteString(")")
	return s.Bytes(), values
}

func (ins *InsertBuilder) Insert(data interface{}) (sql.Result, error) {
	sql, args := ins.parse(data)
	return ins.GetDatabase().ExecPrepare(sql, args...)
}

func (ins *InsertBuilder) InsertM(rows DataSet) {
	for _, r := range rows {
		ins.Insert(r)
	}
}

func (ins *InsertBuilder) Tx(data interface{}) *TxItem {
	sql, args := ins.parse(data)
	return &TxItem{sql, args, nil}
}

func (ins *InsertBuilder) TxM(rows DataSet) []*TxItem {
	arr := make([]*TxItem, 0)
	for _, r := range rows {
		arr = append(arr, ins.Tx(r))
	}
	return arr
}

func (ins *InsertBuilder) LastInsertId(pkey string) int64 {
	database := ins.GetDatabase()
	var count int64 = -1
	r := ins.GetDatabase().Conn.QueryRow(database.Driver.LastInsertId(ins.table, pkey))
	r.Scan(&count)
	return count
}
