package drivers

import (
	"strconv"
)

type Sqlite3 struct {
	File string
}

func (this *Sqlite3) ConnectString() string {
	return this.File
}

func (this *Sqlite3) LimitOffsetStatement(limit int, offset int) string {
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
