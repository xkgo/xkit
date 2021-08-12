package xfile

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/magiconair/properties"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"strings"
)

/**
解析为 kvs map
*/
func ReadAsMap(filePath string) (kvs map[string]string, err error) {
	filename := strings.ToLower(filePath)

	if strings.HasSuffix(filename, "properties") || strings.HasSuffix(filename, "prop") || strings.HasSuffix(filename, "props") {
		return ReadPropertiesAsMap(filePath)
	}

	if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") {
		return ReadYamlAsMap(filePath)
	}
	kvs = make(map[string]string)
	return kvs, errors.New("不支持的 properties 文件类型")
}

/**
读取配置
*/
func ReadPropertiesAsMap(propertiesFile string) (kvs map[string]string, err error) {
	kvs = make(map[string]string)
	file, err := afero.ReadFile(afero.NewOsFs(), propertiesFile)
	if err != nil {
		return nil, err
	}

	in := bytes.NewReader(file)
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(in)
	if err != nil {
		return nil, err
	}

	tempProperties := properties.NewProperties()
	tempProperties.Postfix = ""
	tempProperties.Prefix = ""
	err = tempProperties.Load(buf.Bytes(), properties.UTF8)
	if err != nil {
		return nil, err
	}

	for _, key := range tempProperties.Keys() {
		if val, ok := tempProperties.Get(key); ok {
			kvs[key] = val
		}
	}
	return kvs, nil
}

/**
将 Yaml 文件读取出来，作为 key value 格式
*/
func ReadYamlAsMap(yamlFile string) (kvs map[string]string, err error) {
	kvs = make(map[string]string)
	dataBytes, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return kvs, err
	}

	data := make(map[string]interface{})
	err = yaml.Unmarshal(dataBytes, data)
	if nil != err {
		return kvs, err
	}
	for k, v := range data {
		objectToKvs(k, v, kvs)
	}
	return
}

/**
对象转成 kvs
*/
func objectToKvs(pKey string, obj interface{}, kvs map[string]string) {
	if nil == obj {
		return
	}
	objType := reflect.TypeOf(obj)

	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	if objType.Kind() == reflect.Array || objType.Kind() == reflect.Slice {
		// 数组的话，直接转成json
		dBytes, err := jsoniter.Marshal(obj)
		if nil != err {
			fmt.Println("err:", err)
		}
		kvs[pKey] = string(dBytes)
		return
	}

	// 部署数组，看看是不是map
	if objType.Kind() == reflect.Map {
		objVal := reflect.ValueOf(obj)
		mKeys := objVal.MapKeys()
		for _, mK := range mKeys {
			val := objVal.MapIndex(mK)
			key := fmt.Sprintf("%v", mK.Interface())
			objectToKvs(pKey+"."+key, val.Interface(), kvs)
		}
		return
	}

	// 看看是不是原生对象，Struct 对象
	if objType.Kind() == reflect.Struct {
		objVal := reflect.ValueOf(obj)
		for i := 0; i < objType.NumField(); i++ {
			tfield := objType.Field(i)
			vfield := objVal.Field(i)

			fieldName := tfield.Name
			objectToKvs(pKey+"."+fieldName, vfield.Interface(), kvs)
		}
		return
	}
	// 默认
	kvs[pKey] = fmt.Sprintf("%v", obj)
}
