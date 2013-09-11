# Gos web framework

## db package useage

**begin use db package, you need install and import db driver by your self, like this:**
```go
import(_ "github.com/mattn/go-sqlite3")
```

```go
type DataSet []interface{}
type DataRow map[string]interface{}
```
#### 1. query
```go
q:=&db.QueryBuilder{}   
q.Table("Users").Where("name=? and age=?", "tom", 22).Query() //return []DataRow
q.QueryOne() //return DataRow.(map[string]interface{})   
```

#### 2. query result with struct   
```go
type UserVO struct{
	Name string `db:"name"`
	Age float64 `db:"age"`
	Created time.Time `db:"created_at"`
}
// select name,age,created_at from Usres
q := (&db.QueryBuilder{}).Table("Users")
q.Struct(&UserVO{}) // or
q.Struct(UserVO{})
```
or a nil UserVO pointer
```go
q.Struct((*UserVO)(nil)).QueryOne() //return (*UserVO)
```
#### 3. update
```go
u := (&db.UpdateBuilder{}).Table("Users").Where("id=?", 1)
rowData := db.RowData{}
rowData["name"] = "toms"
u.Update(rowData)
```
or
```go
rowVO = &UserVO{Name:"toms"}
u.Update(rowVO)
```
#### 4. delete
```go
d:=(&db.DeleteBuilder{}).Table("Users").Where("id=?", 1).Delete()
```

#### 5. count
```go
count := (&db.CountBuilder{}).Table("Users").Count()
```

#### 6. db.Query() and db.QueryX()
```go
db.Query("select * from Users") //return []DataRow
```
```go
db.QueryX(&UserVO{}, "select * from Users") //return []*UserVO

```

#### 7. db cache
```go
q.Cache(300).Query() // cache result
```
clear cache
```go
q.ClearCache()
```
**sqlite3**
```go
conf:=make(map[string]string)
conf["driver"] = "sqlite3"
conf["file"] = "./app.db"
// or 
conf["driver"] = "sqlite3"
conf["connect"] = "file=./app.db"
```
**postgres**
```go
conf:=make(map[string]string)
conf["driver"] = "postgres"
conf["dbname"] = "mydb"
conf["host"] = "127.0.0.1"
conf["port"] = "5432"
conf["user"] = "postgres"
conf["password"] = "123"
// or
conf["connect"] = "dbname=mydb user=postgres password=123 host=127.0.0.1 port=5432 sslmode=disable"
```
**init db pool**
```go
db.New("app", conf)
db.New("app2", conf2)
db.Use(0)
```

## log package useage
```go
log.Init("folder", []string{"web", "sql"}, "dev")
log.Level = 10
log.Use("sql")
log.App.Error("error", "code", 2)
// log error and print debug info
log.App.Error("error", "code", 2).Stack()
// auto print debug info
log.PrintStackLevel = 5
```
## conf package useage
config file
```conf
[app]
key=value
key1=value1
key2=value2
key3=value3
key4=value4
```
```go
configuration := conf.Load("app/app.conf")
conf := conf["app"]
boolval  := conf.IsSet("key")
string   := conf.GetString("key1")
intval   := conf.GetInt("key2")
floatval := conf.GetFloat("key3")
boolval  := conf.GetBool("key4")
```
