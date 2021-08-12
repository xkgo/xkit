package xcodec

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"strings"
)

/**
进行MD5转换
*/
func MD5(text string) string {
	hashMaker := md5.New()
	hashMaker.Write([]byte(text))
	return hex.EncodeToString(hashMaker.Sum(nil))
}

func Md5ByBytes(bytes []byte) string {
	hashMaker := md5.New()
	hashMaker.Write(bytes)
	return hex.EncodeToString(hashMaker.Sum(nil))
}

/**
转换成 base64
*/
func EncodeBase64(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

func DecodeBase64(base64Text string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(base64Text)
	if nil != err {
		return "", err
	}
	return string(bytes), nil
}

/**
异或操作
*/
func Xor(text, key string) string {
	keyLen, textLen := len(key), len(text)
	if keyLen < 1 || textLen < 1 {
		return text
	}
	strBuilder := strings.Builder{}
	for i := 0; i < textLen; i++ {
		strBuilder.WriteString(string(text[i] ^ key[i%keyLen]))
	}
	return strBuilder.String()
}

func EncodeBase64XorBase64(text, key string) string {
	// 先进行 base64 操作
	base64Text := EncodeBase64(text)
	xorText := Xor(base64Text, key)
	// 再进行一次Base64(这样子别人看起来就舒服点了)
	return EncodeBase64(xorText)
}

func DecodeBase64XorBase64(text, key string) (string, error) {
	// 先进行一次base64解码
	xorText, err := DecodeBase64(text)
	if nil != err {
		return "", err
	}
	// 先进行 xor 解码
	base64Text := Xor(xorText, key)
	// 再进行一次 base64 解码
	return DecodeBase64(base64Text)
}

/**
获取文件 MD5
*/
func GetFileMd5(filePath string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if nil != err {
		return "", err
	}
	return Md5ByBytes(fileBytes), nil
}
