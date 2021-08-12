package xcodec

import (
	"fmt"
	"testing"
)

func TestEncodeDecodeBase64(t *testing.T) {
	text := "{\"id\":1,\"uid\":123,\"name\":\"Hello哈哈\"}"
	key := "12345678"
	encodeText := EncodeBase64XorBase64(text, key)
	oriText, _ := DecodeBase64XorBase64(encodeText, key)

	fmt.Printf("原文： [%s]\n", text)
	fmt.Printf("密文： [%s]\n", encodeText)
	fmt.Printf("解码： [%s]\n", oriText)
}

func TestDecodeXorBase64(t *testing.T) {
	key := "12345678"
	encodeText := "VEt5RG91fg58YURdUWFbU3hYXEx4XHpLeF8GXFdhYlF+W3l9b2FPS1MZZGBceWBsWHF5DQ=="

	fmt.Println(DecodeBase64XorBase64(encodeText, key))
}

