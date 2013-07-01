package lib

import (
	"encoding/base64"
	"fmt"
	"github.com/jiorry/gos/db"
	"github.com/jiorry/gos/httpd"
	"github.com/jiorry/gos/util"
	"strconv"
	"strings"
	"time"
)

var privateSecret string = util.Unique()

type UserAuth struct {
	login string
	user  db.DataRow
	ctx   *httpd.Context
}

func (this *UserAuth) SetContext(c *httpd.Context) *UserAuth {
	this.ctx = c
	return this
}

func (this *UserAuth) UserAuth(login string, cipher string) bool {
	code, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return false
	}

	now := time.Now()
	ts, _ := strconv.Atoi(fmt.Sprintf("%x", code[0:8]))

	if now.Unix()-int64(ts/1000) > 30 {
		return false
	}
	iv := code[8:24]
	b := AESDecrypt(iv, util.MD5(iv), code[24:])

	user := this.Find(login)
	if user == nil {
		return false
	}

	token := user["Token"].(string)
	if user["Login"] != login || string(b[0:32]) != this.createToken(login, int64(ts), token, "") {
		this.ClearCookie()
		this.login = ""
		this.user = nil
		return false
	}

	this.login = login
	this.user = user
	return true
}

func (this *UserAuth) Find(login string) db.DataRow {
	find := (&db.QueryBuilder{}).Table("Users").Where("login=?", login).Cache(300)

	row, err := find.QueryOne()
	if err != nil {
		return nil
	}

	return row
}

func (this *UserAuth) createToken(login string, ts int64, usertoken string, salt string) string {
	return util.MD5String(fmt.Sprint(login, ts, usertoken, ts, salt))
}

func (this *UserAuth) SetCookie(age int64) {
	ts := time.Now().Unix()
	var unix int64 = 0
	if age > 0 {
		unix = ts + age
	}
	this.ctx.SetCookie("auth", fmt.Sprintf("%s-%d-%s", this.login, ts, this.createToken(this.login, ts, this.user["Token"].(string), privateSecret)), unix, "/", "")
}

func (this *UserAuth) ClearCookie() {
	this.ctx.SetCookie("auth", "", 0, "/", "")
}

func (this *UserAuth) CurrentUser() db.DataRow {
	if this.user != nil {
		return this.user
	}
	v, err := this.ctx.Request.Cookie("auth")
	if err != nil {
		return nil
	}

	arr := strings.Split(v.Value, "-")
	if len(arr) != 3 {
		return nil
	}

	this.login = arr[0]
	this.user = this.Find(this.login)
	n, _ := strconv.Atoi(arr[1])
	ts := int64(n)

	if this.user == nil || arr[2] != this.createToken(this.login, ts, this.user["Token"].(string), privateSecret) {
		this.login = ""
		this.user = nil
	}
	return this.user
}
