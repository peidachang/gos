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

type AuthVO struct {
	Table, FieldId, FieldNick, FieldToken, FieldEmail, FieldSalt, FieldLastSee string
	CookieKey, CookiePublicKey                                                 string
}

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
		if k != nil && k.CreatedAt.Unix() == unix {
			return k
		}
	}

	return nil
}

type UserAuth struct {
	user       db.DataRow
	ctx        *Context
	RegistFunc func(string, string) error
	GroupId    int
	VO         *AuthVO
}

func (this *UserAuth) SetContext(c *Context) *UserAuth {
	this.ctx = c

	if this.VO == nil {
		this.VO = &AuthVO{}
		this.VO.CookieKey = "gosauth"
		this.VO.CookiePublicKey = "gospub"
		this.VO.Table = "users"
		this.VO.FieldId = "id"
		this.VO.FieldNick = "nick"
		this.VO.FieldToken = "token"
		this.VO.FieldEmail = "email"
		this.VO.FieldSalt = "salt"
		this.VO.FieldLastSee = "last_see_at"
	}

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
		return "", NewError(0, "user auth is overdue").Log("notice")
	}

	user := this.Find(login)
	if user == nil {
		return login, NewError(0, "login not found").Log("notice")
	}

	if user.GetString(this.VO.FieldToken) != this.GenerateUserToken(login, pwd, user.GetString(this.VO.FieldSalt)) {
		this.ClearCookie()
		this.user = nil
		return login, NewError(0, "login and password is not matched").Log("notice")
	}

	this.user = user

	data := db.DataRow{}
	data[this.VO.FieldLastSee] = time.Now()
	(&db.UpdateBuilder{}).Table(this.VO.Table).
		Where(this.VO.FieldId+"=?", this.user.GetInt64(this.VO.FieldId)).
		Update(data)

	return login, nil
}

func (this *UserAuth) Find(login string) db.DataRow {
	find := (&db.QueryBuilder{}).Table(this.VO.Table).Where(this.VO.FieldNick+"=?", login).Cache(300)

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

	this.ctx.SetCookie(this.VO.CookieKey, fmt.Sprintf("%s|%d|%s", this.Nick(), ts, this.createAuthToken(this.Nick(), ts, this.user[this.VO.FieldToken].(string), privateSecret)), unix, "/", "", true)
	this.ctx.SetCookie(this.VO.CookiePublicKey, fmt.Sprint(this.Nick(), "|", this.GroupId), unix, "/", "", false)
}

func (this *UserAuth) ClearCookie() {
	this.ctx.SetCookie(this.VO.CookieKey, "", -1, "/", "", true)
	this.ctx.SetCookie(this.VO.CookiePublicKey, "", -1, "/", "", false)
}

func (this *UserAuth) IsOk() bool {
	return len(this.CurrentUser()) > 0
}

func (this *UserAuth) NotOk() bool {
	return len(this.CurrentUser()) == 0
}

func (this *UserAuth) UserId() int64 {
	if this.IsOk() {
		return this.user.GetInt64(this.VO.FieldId)
	}
	return -1
}

func (this *UserAuth) Nick() string {
	if this.IsOk() {
		return this.user.GetString(this.VO.FieldNick)
	}
	return ""
}

func (this *UserAuth) User() db.DataRow {
	return this.CurrentUser()
}

func (this *UserAuth) CurrentUser() db.DataRow {
	if len(this.user) > 0 {
		return this.user
	}
	v, err := this.ctx.Request.Cookie(this.VO.CookieKey)
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

	if this.user == nil || arr[2] != this.createAuthToken(login, ts, this.user[this.VO.FieldToken].(string), privateSecret) {
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
		return 0, nil, NewError(0, "no rsa key found!").Log("error")
	}

	aeskeyBase64, err := rsa.DecryptPKCS1v15(rand.Reader, ppk.Key, rsaCipher)
	if err != nil {
		return 0, nil, NewError(0, "decrypt::", err).Log("error")
	}

	aeskey := make([]byte, 18)
	base64.StdEncoding.Decode(aeskey, aeskeyBase64)
	if len(aeskey) == 0 {
		return 0, nil, NewError(0, "aeskey is empty").Log("error")
	}
	aeskey = aeskey[:16]

	code, err := base64.StdEncoding.DecodeString(string(arr[2]))
	if err != nil {
		return 0, nil, NewError(0, "base64 decode:", err).Log("error")
	}

	now := time.Now()
	ts, _ := strconv.Atoi(fmt.Sprintf("%x", code[0:8]))

	if now.Unix()-int64(ts/1000) > 30 {
		return 0, nil, NewError(0, "login ts is exprie").Log("error")
	}

	b := lib.AESDecrypt(code[8:24], aeskey[0:16], code[24:])

	n := bytes.Index(b, []byte("-"))
	l, err := strconv.Atoi(string(b[0:n]))
	if err != nil {
		return 0, nil, NewError(0, err).Log("error")
	}

	return int64(ts), b[n+1 : n+1+l], nil
}

func (this *UserAuth) Regist(login string, email string, cipher string) error {
	if login == "" || email == "" {
		return NewError(0, "login or email is empty")
	}

	find := (&db.QueryBuilder{}).Table(this.VO.Table)
	if isExist, _ := find.Exists(this.VO.FieldNick+"=? or "+this.VO.FieldEmail+"=?", login, email); isExist {
		return NewError(0, "login or email exists")
	}

	_, text, err := this.PraseCipher([]byte(cipher))
	if err != nil {
		return NewError(0, err)
	}
	salt := util.Unique()
	err = this.RegistFunc(salt, this.GenerateUserToken(login, string(text), salt))

	u := this.Find(login)
	if u == nil {
		return NewError(0, "user is empty")
	} else {
		this.SetUser(u)
		this.SetCookie(0)
	}

	return nil
}
