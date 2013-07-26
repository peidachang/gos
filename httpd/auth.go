package httpd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/jiorry/gos/db"
	"github.com/jiorry/gos/lib"
	"github.com/jiorry/gos/util"
	"strconv"
	"strings"
	"time"
)

var poolRSAKey []*RSAKey
var privateSecret string = util.Unique()
var separator []byte = []byte("|")
var CookieAuthKey string = "bbxauth"
var CookieUserKey string = "bbxuser"

type RSAKey struct {
	Key       *rsa.PrivateKey
	CreatedAt time.Time
}

func init() {
	poolRSAKey = make([]*RSAKey, 3)
}

func newRSAKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return nil
	}
	return key
}

// The Rsa key will be changed after every 3 minute
func GetRSAKey(unix int64) *RSAKey {
	now := time.Now()
	if poolRSAKey[0] == nil {
		poolRSAKey[0] = &RSAKey{newRSAKey(), now}
	} else {
		if now.Unix()-poolRSAKey[0].CreatedAt.Unix() > 180 {
			poolRSAKey[2] = poolRSAKey[1]
			poolRSAKey[1] = poolRSAKey[0]
			poolRSAKey[0] = &RSAKey{newRSAKey(), now}
		}
	}

	if unix == 0 {
		return poolRSAKey[0]
	}

	for _, k := range poolRSAKey {
		if k.CreatedAt.Unix() == unix {
			return k
		}
	}

	return nil
}

type UserAuth struct {
	user db.DataRow
	ctx  *Context
}

func (this *UserAuth) SetContext(c *Context) *UserAuth {
	this.ctx = c
	return this
}

func (this *UserAuth) SetUser(row db.DataRow) *UserAuth {
	this.user = row
	return this
}

func (this *UserAuth) GenerateUserToken(login string, pwd string, salt string) string {
	return util.MD5String(login + salt + pwd + salt)
}

func (this *UserAuth) Auth(cipher []byte) (string, error) {
	ts, b, err := this.PraseCipher(cipher)
	if err != nil {
		return "", err
	}
	arr := bytes.Split(b, separator)
	login := string(arr[0])
	pwd := string(arr[1])

	if time.Now().Unix()-int64(ts) > 30 {
		return "", MyErr(0, "user auth is overdue").Log("notice")
	}

	user := this.Find(login)
	if user == nil {
		return login, MyErr(0, "login name not found").Log("notice")
	}

	if user.GetString("Token") != this.GenerateUserToken(login, pwd, user.GetString("Salt")) {
		this.ClearCookie()
		this.user = nil
		return login, MyErr(0, "login name password is not matched").Log("notice")
	}

	this.user = user
	return login, nil
}

func (this *UserAuth) Find(login string) db.DataRow {
	find := (&db.QueryBuilder{}).Table("Users").Where("login=?", login).Cache(300)

	row, err := find.QueryOne()
	if err != nil {
		return nil
	}

	return row
}

func (this *UserAuth) createAuthToken(login string, ts int64, usertoken string, salt string) string {
	return util.MD5String(fmt.Sprint(salt, login, ts, usertoken, salt))
}

func (this *UserAuth) SetCookie(age int64) {
	ts := time.Now().Unix()
	var unix int64 = 0
	if age > 0 {
		unix = ts + age
	}

	this.ctx.SetCookie(CookieAuthKey, fmt.Sprintf("%s|%d|%s", this.UserName(), ts, this.createAuthToken(this.UserName(), ts, this.user["Token"].(string), privateSecret)), unix, "/", "", true)
	this.ctx.SetCookie(CookieUserKey, this.UserName(), unix, "/", "", false)
}

func (this *UserAuth) ClearCookie() {
	this.ctx.SetCookie(CookieAuthKey, "", -1, "/", "", true)
	this.ctx.SetCookie(CookieUserKey, "", -1, "/", "", false)
}

func (this *UserAuth) IsLogin() bool {
	return len(this.CurrentUser()) > 0
}

func (this *UserAuth) IsNotLogin() bool {
	return len(this.CurrentUser()) == 0
}

func (this *UserAuth) UserId() int64 {
	if this.IsLogin() {
		return this.user["Id"].(int64)
	}
	return -1
}

func (this *UserAuth) UserName() string {
	if this.IsLogin() {
		return this.user["Login"].(string)
	}
	return ""
}

func (this *UserAuth) CurrentUser() db.DataRow {
	if len(this.user) > 0 {
		return this.user
	}
	v, err := this.ctx.Request.Cookie(CookieAuthKey)
	if err != nil {
		return this.user
	}

	arr := strings.Split(v.Value, string(separator))
	if len(arr) != 3 {
		return this.user
	}

	login := arr[0]
	this.user = this.Find(login)
	n, _ := strconv.Atoi(arr[1])
	ts := int64(n)

	if this.user == nil || arr[2] != this.createAuthToken(login, ts, this.user["Token"].(string), privateSecret) {
		this.user = nil
	}

	return this.user
}

func (this *UserAuth) PraseCipher(cipher []byte) (int64, []byte, error) {
	arr := bytes.Split(cipher, separator)
	rsakeyUnix, _ := strconv.Atoi(string(arr[0]))

	rsaCipher := make([]byte, len(arr[1])/2)
	hex.Decode(rsaCipher, arr[1])

	ppk := GetRSAKey(int64(rsakeyUnix))
	if ppk == nil {
		return 0, nil, MyErr(0, "no rsa key found!").Log("error")
	}

	aeskeyBase64, err := rsa.DecryptPKCS1v15(rand.Reader, ppk.Key, rsaCipher)
	if err != nil {
		return 0, nil, MyErr(0, "decrypt::", err).Log("error")
	}

	aeskey := make([]byte, 18)
	base64.StdEncoding.Decode(aeskey, aeskeyBase64)
	if len(aeskey) == 0 {
		return 0, nil, MyErr(0, "aeskey is empty").Log("error")
	}
	aeskey = aeskey[:16]

	code, err := base64.StdEncoding.DecodeString(string(arr[2]))
	if err != nil {
		return 0, nil, MyErr(0, "base64 decode:", err).Log("error")
	}

	now := time.Now()
	ts, _ := strconv.Atoi(fmt.Sprintf("%x", code[0:8]))

	if now.Unix()-int64(ts/1000) > 30 {
		return 0, nil, MyErr(0, "login ts is exprie").Log("error")
	}

	b := lib.AESDecrypt(code[8:24], aeskey[0:16], code[24:])

	n := bytes.Index(b, []byte("-"))
	l, err := strconv.Atoi(string(b[0:n]))
	if err != nil {
		return 0, nil, MyErr(0, err).Log("error")
	}

	return int64(ts), b[n+1 : n+1+l], nil
}
