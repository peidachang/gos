package drivers

import (
	"fmt"
)

type Postgres struct {
	Common
	Dbname   string
	User     string
	Password string
	Host     string
	Port     string
}

func (this *Postgres) ConnectString() string {
	host := "localhost"
	if this.Host != "" {
		host = this.Host
	}
	port := "5432"
	if this.Port != "" {
		port = this.Port
	}
	return fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%s sslmode=disable",
		this.Dbname,
		this.User,
		this.Password,
		host,
		port)
}

func (this *Postgres) QuoteField(s string) string {
	return `"` + s + `"`
}

func (this *Postgres) LastInsertId(table, pkey string) string {
	return "select currval(pg_get_serial_sequence('" + table + "','" + pkey + "'))"
}
