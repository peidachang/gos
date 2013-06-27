package drivers

import (
	"fmt"
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
