package db

import (
	"database/sql"
	"github.com/jiorry/gos/log"
)

type TxItem struct {
	SqlCode    []byte
	Args       []interface{}
	DataStruct *structMaps
}

type Tx struct {
	builder
	tx        *sql.Tx
	IsError   bool
	LastError error
}

func (t *Tx) GetTx() *sql.Tx {
	if t.tx != nil {
		return t.tx
	}

	tx, err := t.GetDatabase().Conn.Begin()
	if err != nil {
		log.App.Alert(err)
		return nil
	}
	t.tx = tx
	t.IsError = false
	return t.tx
}

func (t *Tx) Query(item *TxItem) (DataSet, error) {
	if t.IsError {
		return nil, nil
	}
	sqlstr := string(t.GetDatabase().AdaptSql(item.SqlCode))
	dblog.Sql(sqlstr, item.Args)
	rows, err := t.GetTx().Query(sqlstr, item.Args...)
	if t.isError(err) {
		return nil, err
	}
	defer rows.Close()

	dataset, err := scanRows(item.DataStruct, rows)
	if t.isError(err) {
		return nil, err
	}

	return dataset, nil
}

func (t *Tx) QueryPrepare(item *TxItem) (DataSet, error) {
	if t.IsError {
		return nil, nil
	}

	sqlstr := string(t.GetDatabase().AdaptSql(item.SqlCode))
	dblog.Sql(sqlstr, item.Args)
	s, err := t.GetTx().Prepare(sqlstr)
	if t.isError(err) {
		return nil, err
	}

	rows, err := s.Query(item.Args...)
	if t.isError(err) {
		return nil, err
	}

	defer rows.Close()

	dataset, err := scanRows(item.DataStruct, rows)
	if t.isError(err) {
		return nil, err
	}

	return dataset, nil
}

func (t *Tx) Exec(item *TxItem) (sql.Result, error) {
	if t.IsError {
		return nil, nil
	}
	sqlstr := string(t.GetDatabase().AdaptSql(item.SqlCode))
	dblog.Sql(sqlstr, item.Args)
	r, err := t.GetTx().Exec(sqlstr, item.Args...)
	if t.isError(err) {
		return nil, err
	}
	return r, nil
}

func (t *Tx) LastInsertId(table, pkey string) (int64, error) {
	if t.IsError {
		return -1, nil
	}
	r := t.tx.QueryRow(t.GetDatabase().Driver.LastInsertId(table, pkey))

	var count int64
	err := r.Scan(&count)
	if t.isError(err) {
		return -1, err
	}
	return count, err
}

func (t *Tx) ExecPrepare(item *TxItem) (sql.Result, error) {
	if t.IsError {
		return nil, nil
	}
	sqlstr := string(t.GetDatabase().AdaptSql(item.SqlCode))
	dblog.Sql(sqlstr, item.Args)
	s, err := t.GetTx().Prepare(sqlstr)
	if t.isError(err) {
		return nil, err
	}
	r, err := s.Exec(item.Args...)
	if t.isError(err) {
		return nil, err
	}
	return r, nil
}

func (t *Tx) Do(items ...*TxItem) {
	if t.IsError {
		return
	}
	for _, item := range items {
		_, err := t.ExecPrepare(item)
		if err != nil {
			return
		}
	}
	t.tx.Commit()
}

func (t *Tx) Commit() error {
	if t.IsError {
		return nil
	}
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	if t.IsError {
		return nil
	}
	return t.tx.Rollback()
}

func (t *Tx) isError(err error) bool {
	if err != nil {
		log.App.Error(err)
		t.tx.Rollback()
		t.IsError = true
		t.LastError = err
		return true
	}
	return false
}
