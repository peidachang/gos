package drivers

import (
	"fmt"
)

type Mysql struct {
	Common
	Dbname   string
	User     string
	Password string
	Host     string
	Port     string
	Charset  string
}

func (this *Mysql) ConnectString() string {
	charset := "utf8"
	if this.Charset != "" {
		charset = this.Charset
	}
	host := "localhost"
	if this.Host != "" {
		host = this.Host
	}
	port := "3306"
	if this.Port != "" {
		port = this.Port
	}

	return fmt.Sprintf("?:?@tcp(?:?)/??charset=?",
		this.User,
		this.Password,
		host,
		port,
		this.Dbname,
		charset,
	)
}

func (this *Mysql) QuoteField(s string) string {
	return `"` + s + `"`
}

func (this *Mysql) LastInsertId(table, id string) string {
	return "SELECT LAST_INSERT_ID()"
}
