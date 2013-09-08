package db

import (
	"bytes"
	"encoding/gob"
	"github.com/jiorry/gos/cache"
	"github.com/jiorry/gos/log"
	"time"
)

func init() {
	gob.Register(DataRow{})
	gob.Register(DataSet{})
	gob.Register(time.Time{})
}

type builder struct {
	database *Database
}

func (this *builder) GetDatabase() *Database {
	if this.database == nil {
		this.database = Current()
	}
	return this.database
}

func (this *builder) SetDatabase(d *Database) {
	this.database = d
}

type parpareParams struct {
	code string
	args []interface{}
}

func keyValueList(data DataRow) (keys []string, values []interface{}, stmts []string) {
	length := len(data)
	keys = make([]string, length)
	values = make([]interface{}, length)
	stmts = make([]string, length)
	i := 0
	for k, v := range data {
		keys[i] = k
		stmts[i] = "?"
		values[i] = v
		i++
	}

	return
}

func cacheSet(key string, value interface{}, expire int) error {
	if !cache.IsEnable() {
		return nil
	}

	if !cache.IsEnable() {
		return nil
	}
	v, err := gobEncode(value)
	if err != nil {
		log.App.Crit(err)
		return err
	}
	err = cache.Set(key, v, expire)
	if err != nil {
		log.App.Crit(err)
	}
	return err
}

func cacheGetDBResult(key string) (DataSet, error) {
	out := DataSet{}
	reply, err := cache.Get(key)
	if reply == nil || err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(reply.([]byte)))
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func cacheGetDBRow(key string) (DataRow, error) {
	var out = DataRow{}
	reply, err := cache.Get(key)
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(reply.([]byte)))
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func gobEncode(obj interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return []byte(""), err
	}
	return buf.Bytes(), nil
}
