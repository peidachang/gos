# Gos web framework

## db package useage
```go
type DataSet []interface{}
type DataRow map[string]interface{}
```
## 1. query
```go
q:=&db.QueryBuilder{}   
q.Table("Users").Where("name=? and age=?", "tom", 22).Query()** //DataSet is return   
q->QueryOne()** //DataRow is return   
```

## 2. query and return DataSet with Struct Row Data   
```go
type UserVO struct{
	Name string \`db:"name"\`
	Age float64 \`db:"age"\`
	Created time.Time \`db:"created_at"\`
}
q := (&db.QueryBuilder{}).Table("Users")
q.Struct(&UserVO{}) || q.Struct(UserVO{})
```
or a nil UserVO pointer
```go
q.Struct((*UserVO)(nil)).QueryOne()** //(*UserVO) is return
```
## 3. update
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
4. delete   
```go
d:=(&db.DeleteBuilder{}).Table("Users").Where("id=?", 1).Delete()
```

5. count   
```go
count := (&db.CountBuilder{}).Table("Users").Count()
```