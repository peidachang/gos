package util

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"hash/crc64"
	"time"
)

func NowString() string {
	return time.Now().Format("2006-01-02 03:04:05")
}

func Now() int64 {
	return time.Now().Unix()
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
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
