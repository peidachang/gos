package drivers

import (
	"fmt"
	"strconv"
)

type Postgres struct {
	Sqlite3
	Dbname   string
	User     string
	Password string
	Host     string
	Port     string
}

func (this *Postgres) ConnectString() string {
	return fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%s sslmode=disable",
		this.Dbname,
		this.User,
		this.Password,
		this.Host,
		this.Port)
}

func (this *Postgres) LimitOffsetStatement(limit int, offset int) string {
	if !(limit > 0 || offset > 0) {
		return ""
	}

	if limit > 0 && offset > 0 {
		return " limit " + strconv.Itoa(limit) + " offset " + strconv.Itoa(offset)
	} else if limit > 0 {
		return " limit " + strconv.Itoa(limit)
	} else {
		return " offset " + strconv.Itoa(offset)
	}
}

func (this *Postgres) QuoteField(s string) string {
	return `"` + s + `"`
}
