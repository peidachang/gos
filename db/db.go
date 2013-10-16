package db

import (
	"bytes"
	"database/sql"
	"github.com/jiorry/gos/conf"
	"github.com/jiorry/gos/db/drivers"
	"github.com/jiorry/gos/log"
	"io/ioutil"
)

type databasePool struct {
	dblist  map[string]*Database
	current *Database
}

func (this *databasePool) Current() *Database {
	return this.current
}
func (this *databasePool) SetCurrent(d *Database) {
	this.current = d
}
func (this *databasePool) Use(name string) {
	c := this.GetDatabase(name)
	if c == nil {
		log.App.Emerg(name, " database is not found!")
		return
	} else {
		this.current = c
	}
}
func (this *databasePool) SetDatabase(name string, d *Database) {
	this.dblist[name] = d
}
func (this *databasePool) GetDatabase(name string) *Database {
	if v, ok := this.dblist[name]; ok {
		return v
	}
	return nil
}

var (
	dbpool *databasePool
	dblog  *log.Logger
)

var commaSplit []byte = []byte(",")

func init() {
	dblog = nil
	dbpool = &databasePool{dblist: make(map[string]*Database)}
}

func Init(name string, conf conf.Conf) {
	dbpool.SetCurrent(New(name, conf))
}

// create a database instance
func New(name string, conf conf.Conf) *Database {
	if dbpool.GetDatabase(name) != nil {
		log.App.Alert(name, "this database is already exists!")
		return nil
	}

	if conf == nil {
		conf = map[string]string{}
		conf["driver"] = "sqlite3"
		conf["file"] = "./app.db"
	}
	d := &Database{}
	d.Name = name
	d.DriverName = conf.Get("driver")

	switch d.DriverName {
	case "postgres":
		d.Driver = &drivers.Postgres{Dbname: conf.Get("dbname"),
			User:     conf.Get("user"),
			Password: conf.Get("password"),
			Host:     conf.Get("host"),
			Port:     conf.Get("port")}
	case "mysql":
		d.Driver = &drivers.Mysql{Dbname: conf.Get("dbname"),
			User:     conf.Get("user"),
			Password: conf.Get("password"),
			Host:     conf.Get("host"),
			Port:     conf.Get("port"),
			Charset:  conf.Get("charset")}
	case "sqlite3":
		d.Driver = &drivers.Sqlite3{File: conf.Get("file")}
	default:
		log.App.Notice("you may need regist a custom driver: db.RegistDriver(Mysql{})")
		d.Driver = &drivers.Common{}
	}
	d.Driver.SetConnectString(conf.Get("connect"))
	d.Connect()

	if dblog == nil {
		dblog = log.New("db")
	}

	dbpool.SetDatabase(name, d)

	return d
}

// Query from database on prepare mode
// This function use the current database from database bool
// You can set another database use Use(i) or create an new database use New(name, conf)
func Query(sqlstr string, args ...interface{}) (DataSet, error) {
	return Current().Query([]byte(sqlstr), args...)
}

// Query from database on prepare mode
// In prepare mode, the sql command will be cached by database
// This function use the current database from database bool
// You can set another database by Use(i) or New(name, conf) an new database
func QueryPrepare(sqlstr string, args ...interface{}) (DataSet, error) {
	return Current().QueryPrepare([]byte(sqlstr), args...)
}

func QueryX(cls interface{}, sqlstr string, args ...interface{}) (DataSet, error) {
	return Current().QueryX(cls, []byte(sqlstr), args...)
}
func QueryPrepareX(cls interface{}, sqlstr string, args ...interface{}) (DataSet, error) {
	return Current().QueryPrepareX(cls, []byte(sqlstr), args...)
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
	return Current().Exec([]byte(sqlstr), args...)
}

// Excute sql on prepare mode
// This function use the current database from database bool
func ExecPrepare(sqlstr string, args ...interface{}) (sql.Result, error) {
	return Current().ExecPrepare([]byte(sqlstr), args...)
}

// Get a database instance by name from database pool
func Get(name string) *Database {
	return dbpool.GetDatabase(name)
}

// Return the current database from database pool
func Current() *Database {
	return dbpool.Current()
}

// Set current database by index
func Use(name string) {
	dbpool.Use(name)
}

// Get database count
func DatabaseCount() int {
	return len(dbpool.dblist)
}
