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
	"github.com/jiorry/gos/log"
	"github.com/jiorry/gos/util"
	"strconv"
	"strings"
	"time"
)

var poolRSAKey []*RSAKey
var privateSecret string = util.Unique()

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

func (this *UserAuth) UserAuth(login string, cipher []byte) bool {
	ts, b, err := this.PraseCipher(cipher)
	if err != nil {
		return false
	}

	user := this.Find(login)
	if user == nil {
		log.App.Notice("login name not found:", login)
		return false
	}

	token := user["Token"].(string)
	if user["Login"] != login || fmt.Sprintf("%x", b[0:16]) != this.createToken(login, int64(ts), token, "") {
		// fmt.Printf("%s-%s-\n%s\n%s\n", user["Login"], login, fmt.Sprintf("%x", b[0:16]), this.createToken(login, int64(ts), token, ""))
		this.ClearCookie()
		this.user = nil
		return false
	}

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

	this.ctx.SetCookie("auth", fmt.Sprintf("%s-%d-%s", this.UserName(), ts, this.createToken(this.UserName(), ts, this.user["Token"].(string), privateSecret)), unix, "/", "")
}

func (this *UserAuth) ClearCookie() {
	this.ctx.SetCookie("auth", "", -1, "/", "")
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
	v, err := this.ctx.Request.Cookie("auth")
	if err != nil {
		return this.user
	}

	arr := strings.Split(v.Value, "-")
	if len(arr) != 3 {
		return this.user
	}

	login := arr[0]
	this.user = this.Find(login)
	n, _ := strconv.Atoi(arr[1])
	ts := int64(n)

	if this.user == nil || arr[2] != this.createToken(login, ts, this.user["Token"].(string), privateSecret) {
		this.user = nil
	}

	return this.user
}

func (this *UserAuth) PraseCipher(cipher []byte) (int64, []byte, error) {
	arr := bytes.Split(cipher, []byte(" "))
	rsakeyUnix, _ := strconv.Atoi(string(arr[0]))

	rsaCipher := make([]byte, len(arr[1])/2)
	hex.Decode(rsaCipher, arr[1])

	ppk := GetRSAKey(int64(rsakeyUnix))
	if ppk == nil {
		return 0, nil, MyErr(0, "no rsa key found!").Log()
	}

	aeskeyBase64, err := rsa.DecryptPKCS1v15(rand.Reader, ppk.Key, rsaCipher)
	if err != nil {
		return 0, nil, MyErr(0, "decrypt::", err).Log()
	}

	aeskey := make([]byte, 18)
	base64.StdEncoding.Decode(aeskey, aeskeyBase64)
	if len(aeskey) == 0 {
		return 0, nil, MyErr(0, "aeskey is empty").Log()
	}
	aeskey = aeskey[:16]

	code, err := base64.StdEncoding.DecodeString(string(arr[2]))
	if err != nil {
		return 0, nil, MyErr(0, "base64 decode:", err).Log()
	}

	now := time.Now()
	ts, _ := strconv.Atoi(fmt.Sprintf("%x", code[0:8]))

	if now.Unix()-int64(ts/1000) > 30 {
		return 0, nil, MyErr(0, "login ts is exprie").Log()
	}

	b := lib.AESDecrypt(code[8:24], aeskey[0:16], code[24:])
	return int64(ts), b, nil
}
