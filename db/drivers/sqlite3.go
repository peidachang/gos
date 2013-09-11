package drivers

type Sqlite3 struct {
	Common
	File string
}

func (this *Sqlite3) ConnectString() string {
	return this.File
}

func (this *Sqlite3) QuoteField(s string) string {
	return `"` + s + `"`
}
