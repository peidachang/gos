package db

import (
	"bytes"
	"database/sql"
	"github.com/jiorry/gos/db/drivers"
	"github.com/jiorry/gos/log"
	"io/ioutil"
)

type databasePool struct {
	dblist       []*Database
	currentIndex int
}

var (
	dbpool *databasePool
	dblog  *log.Logger
)

var commaSplit []byte = []byte(",")

func init() {
	dblog = nil
	dbpool = &databasePool{dblist: make([]*Database, 0)}
}

// create a database instance
func New(name string, mapData map[string]string) {
	if mapData == nil {
		mapData = map[string]string{}
		mapData["driver"] = "sqlite3"
		mapData["file"] = "./app.db"
	}
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

	if dblog == nil {
		dblog = log.Add("db")
	}

	Add(d)
	Use(DatabaseCount() - 1)
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

func QueryX(cls interface{}, sqlstr string, args ...interface{}) (DataSet, error) {
	return Current().QueryX(cls, sqlstr, args...)
}
func QueryPrepareX(cls interface{}, sqlstr string, args ...interface{}) (DataSet, error) {
	return Current().QueryPrepareX(cls, sqlstr, args...)
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
