package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"math/big"
	"strconv"
)

//字符串转数字
func Atoi(s string, defaultValue int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return i
}

//生成随机字符串
//l为字符串长度
//t为类型默认0  0-大小写字母字符串 1-数字字符串 2-大小写字母+数字字符串
func RandString(l int, t ...int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if len(t) > 0 {
		switch t[0] {
		case 1:
			str = "0123456789"
		case 2:
			str = str + "0123456789"
		default:
			break
		}
	}
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < l; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}

//md5字符串
func Md5String(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return hex.EncodeToString(m.Sum(nil))
}

//sha1 字符串
func Sha1String(str string) string {
	s := sha1.New()
	s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}