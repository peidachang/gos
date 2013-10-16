package drivers

type Common struct {
	connect string
}

func (c *Common) SetConnectString(s string) {
	c.connect = s
}

func (c *Common) ConnectString() string {
	return c.connect
}

func (c *Common) QuoteField(s string) string {
	return `"` + s + `"`
}

func (c *Common) LastInsertId(table, pkey string) string {
	return ""
}
