package drivers

type Common struct {
	connect string
}

func (this *Common) SetConnectString(s string) {
	this.connect = s
}

func (this *Common) ConnectString() string {
	return this.connect
}

func (this *Common) QuoteField(s string) string {
	return `"` + s + `"`
}
