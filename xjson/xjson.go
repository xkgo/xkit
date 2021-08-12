package xjson

import (
	"fmt"
	"github.com/json-iterator/go"
)

func ToJsonString(v interface{}) (content string, err error) {
	bytes, err := jsoniter.Marshal(v)

	if err == nil {
		content = string(bytes)
	}
	return
}

func ToJsonStringWithoutError(v interface{}) string {
	bytes, err := jsoniter.Marshal(v)

	if err == nil {
		return string(bytes)
	}
	return fmt.Sprint(v)
}

func FromJson(jsonStr string, v interface{}) (interface{}, error) {
	err := jsoniter.Unmarshal([]byte(jsonStr), v)
	return v, err
}
