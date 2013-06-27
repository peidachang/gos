package db

import (
	"bytes"
	"database/sql"
	"github.com/jiorry/yundata.com/db/drivers"
	"github.com/jiorry/yundata.com/log"
	"io/ioutil"
	"strings"
	"time"
)

type databasePool struct {
	dblist       []*Database
	currentIndex int
}

type IDriver interface {
	ConnectString() string
	LimitOffsetStatement(int, int) string
}

type Database struct {
	Name       string
	Driver     IDriver
	DriverName string
	Conn       *sql.DB
}

var (
	IsWriteLog bool

	dbpool *databasePool
	logger *log.Logger
)

func init() {
	logger = nil
	dbpool = &databasePool{dblist: make([]*Database, 0)}
}

// create a database instance
func New(name string, mapData map[string]string) {
	d := &Database{}
	d.Name = name
	d.DriverName = mapData["driver"]
	switch d.DriverName {
	case "postgres":
		d.Driver = &drivers.Postgres{Dbname: mapData["dbname"],
			User:     mapData["user"],
			Password: mapData["password"],
			Host:     mapData["host"],
			Port:     mapData["port"]}
	case "sqlite3":
		d.Driver = &drivers.Sqlite3{File: mapData["file"]}
	case "none":
		return
	default:
		log.App.Fatalln("no db driver found. you may choose the follows: sqlite3, mysql, postgres, none")
	}

	d.Connect()

	if IsWriteLog && logger == nil {
		logger = log.Add("db")
	}

	Add(d)
	Use(DatabaseCount() - 1)
}

// Connect database
func (this *Database) Connect() error {
	c, err := sql.Open(this.DriverName, this.Driver.ConnectString())
	if err != nil {
		log.App.Emerg(err)
		return err
	}
	this.Conn = c
	return nil
}

// Excute query on db prepare mode
func (this *Database) QueryPrepare(sqlstr string, args ...interface{}) (DataSet, error) {
	s, _ := this.Conn.Prepare(sqlstr)
	rows, err := s.Query(args...)
	if err != nil {
		log.App.Alert(err)
		log.App.Stack()
		return nil, err
	}
	defer rows.Close()

	var result DataSet
	result, err = parseRow(rows)
	if err != nil {
		return nil, err
	}
	if IsWriteLog {
		logger.Sql(sqlstr, args)
	}

	return result, nil
}

// Query from database, return DataSet result collection.
func (this *Database) Query(sqlstr string, args ...interface{}) (DataSet, error) {
	if len(strings.TrimSpace(sqlstr)) == 0 {
		return nil, nil
	}

	var rows *sql.Rows
	var err error
	rows, err = this.Conn.Query(sqlstr, args...)

	if err != nil {
		log.App.Alert(err)
		log.App.Stack()
		return nil, err
	}
	defer rows.Close()

	var result DataSet
	result, err = parseRow(rows)
	if err != nil {
		return nil, err
	}
	if IsWriteLog {
		logger.Sql(sqlstr, args)
	}

	return result, nil
}

// Excute sql command on db prepare mode
// In prepare mode, the sql command will be cached by database
func (this *Database) ExecPrepare(sqlstr string, args ...interface{}) (sql.Result, error) {
	s, _ := this.Conn.Prepare(sqlstr)
	r, err := s.Exec(args...)
	if err != nil {
		log.App.Alert("db exec error:", err, "\n", "sql:"+sqlstr+"|")
		log.App.Stack()
		return nil, err
	}
	return r, nil
}

// Excute sql.
// If your has more than on sql command, it will only excute the first.
func (this *Database) Exec(sqlstr string, args ...interface{}) (sql.Result, error) {
	r, err := this.Conn.Exec(sqlstr, args...)
	if err != nil {
		log.App.Alert("db exec error:", err, "\n", "sql:"+sqlstr+"|")
		log.App.Stack()
		return nil, err
	}
	return r, nil
}

// Query from database on prepare mode
// This function use the current database from database bool
// You can set another database use Use(i) or create an new database use New(name, conf)
func Query(sqlstr string, args ...interface{}) (DataSet, error) {
	return Current().Query(sqlstr, args...)
}

// Query from database on prepare mode
// In prepare mode, the sql command will be cached by database
// This function use the current database from database bool
// You can set another database by Use(i) or New(name, conf) an new database
func QueryPrepare(sqlstr string, args ...interface{}) (DataSet, error) {
	return Current().QueryPrepare(sqlstr, args...)
}

// Excute sql from a file
// This function run sql under not transaction mode and use the current database from database bool
func ExecFromFile(file string) error {
	var filebytes []byte
	var err error

	if filebytes, err = ioutil.ReadFile(file); err != nil {
		return err
	}

	b := bytes.Split(filebytes, []byte(";"))
	for _, i := range b {
		if len(bytes.TrimSpace(i)) == 0 {
			continue
		}

		_, err = Exec(string(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// Excute sql.
// If your has more than on sql command, it will only excute the first.
// This function use the current database from database bool
func Exec(sqlstr string, args ...interface{}) (sql.Result, error) {
	return Current().Exec(sqlstr, args...)
}

// Excute sql on prepare mode
// This function use the current database from database bool
func ExecPrepare(sqlstr string, args ...interface{}) (sql.Result, error) {
	return Current().ExecPrepare(sqlstr, args...)
}

// Add a database to database pool
func Add(d *Database) {
	dbpool.dblist = append(dbpool.dblist, d)
}

// Get a database instance by name from database pool
func Get(name string) *Database {
	for _, item := range dbpool.dblist {
		if name == item.Name {
			return item
		}
	}
	return nil
}

// Get a database instance by index from database pool
func GetByIndex(i int) *Database {
	return dbpool.dblist[i]
}

// Return the current database from database pool
func Current() *Database {
	return dbpool.dblist[dbpool.currentIndex]
}

// Set current database by index
func Use(i int) *Database {
	dbpool.currentIndex = i
	return Current()
}

// Get database count
func DatabaseCount() int {
	return len(dbpool.dblist)
}

type DataRow map[string]interface{}
type DataSet []DataRow

func (this DataRow) GetString(field string) string {
	return this[field].(string)
}

func (this DataRow) GetInt(field string) int64 {
	return this[field].(int64)
}

func (this DataRow) GetFloat(field string) float64 {
	return this[field].(float64)
}

func (this DataRow) GetTime(field string) time.Time {
	return this[field].(time.Time)
}

func (this DataSet) Search(field string, value interface{}) DataRow {
	for _, row := range this {
		if row[field] == value {
			return row
		}
	}
	return nil
}

func (this DataSet) Encode() [][]interface{} {
	var value []interface{}
	var colSize int

	values := make([][]interface{}, len(this)+1)
	columns := make([]interface{}, 0)
	isFirst := true

	for i, row := range this {
		if isFirst {
			for k, _ := range row {
				columns = append(columns, k)
			}
			colSize = len(columns)
			values[i] = columns
			isFirst = false
		}

		value = make([]interface{}, colSize)
		n := 0
		for _, v := range row {
			value[n] = v
			n++
		}

		values[i+1] = value
	}

	return values
}

func parseRow(rows *sql.Rows) (DataSet, error) {
	cols, _ := rows.Columns()
	colsNum := len(cols)

	result := DataSet{}
	var err error
	var row, tem []interface{}
	var rowData map[string]interface{}

	for rows.Next() {
		row = make([]interface{}, colsNum)
		tem = make([]interface{}, colsNum)

		for i := range row {
			tem[i] = &row[i]
		}

		if err = rows.Scan(tem...); err != nil {
			log.App.Err(err)
			return nil, err
		}

		rowData = make(map[string]interface{})
		for i := range cols {
			rowData[cols[i]] = row[i]
		}

		result = append(result, rowData)
	}

	if err = rows.Err(); err != nil {
		log.App.Err(err)
		log.App.Stack()
		return nil, err
	}
	return result, nil
}
