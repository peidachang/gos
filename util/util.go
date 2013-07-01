package util

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash/crc64"
	"io"
	"time"
)

func NowString() string {
	return time.Now().Format("2006-01-02 03:04:05")
}

func MD5String(str string) string {
	return fmt.Sprintf("%x", MD5([]byte(str)))
}

func MD5(b []byte) []byte {
	h := md5.New()
	h.Write(b)
	return h.Sum(nil)
}

func Sha1(str string) string {
	h := sha1.New()
	io.WriteString(h, "abc")
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Unique() string {
	nowtime := time.Now()
	t := nowtime.UnixNano()
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, t)

	tab := crc64.MakeTable(uint64(nowtime.Unix()))
	i := crc64.Checksum(buf.Bytes(), tab)
	return string(IntToBaseN(i))
}

func IntToBaseN(num uint64) []byte {
	l := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	length := uint64(len(l))

	parse := func(n uint64) (v uint64, m uint64) {
		return n / length, n % length
	}
	result := []byte{}
	v, m := parse(num)
	result = append(result, l[m])

	for v >= length {
		v, m = parse(v)
		result = append([]byte{byte(l[m])}, result...)
	}
	if v > 0 {
		result = append([]byte{byte(l[v])}, result...)
	}
	return result
}

func InStringArray(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
