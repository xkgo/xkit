package xreflect

import (
	"encoding/json"
	"github.com/xkgo/xkit/xconv"
	"github.com/xkgo/xkit/xstr"
	"reflect"
	"strings"
)

var (
	DefString  string  = ""
	DefUint    uint    = 0
	DefUint8   uint8   = 0
	DefUint16  uint16  = 0
	DefUint32  uint32  = 0
	DefUint64  uint64  = 0
	DefInt     int     = 0
	DefInt8    int8    = 0
	DefInt16   int16   = 0
	DefInt32   int32   = 0
	DefInt64   int64   = 0
	DefFloat32 float32 = 0
	DefFloat64 float64 = 0
	DefBool    bool    = false
	DefRune    rune    = 0 // int32 别称
	DefByte    byte    = 0 // uint8 别称
)

var (
	ErrorType    = newErrorReflectType()
	ErrorTypePtr = reflect.PtrTo(newErrorReflectType())

	String    = newReflectType(DefString)
	StringPtr = newReflectType(&DefString)

	Uint    = newReflectType(DefUint)
	UintPtr = newReflectType(&DefUint)

	Uint8    = newReflectType(DefUint8)
	Uint8Ptr = newReflectType(&DefUint8)

	Uint16    = newReflectType(DefUint16)
	Uint16Ptr = newReflectType(&DefUint16)

	Uint32    = newReflectType(DefUint32)
	Uint32Ptr = newReflectType(&DefUint32)

	Uint64    = newReflectType(DefUint64)
	Uint64Ptr = newReflectType(&DefUint64)

	Int    = newReflectType(DefInt)
	IntPtr = newReflectType(&DefInt)

	Int8    = newReflectType(DefInt8)
	Int8Ptr = newReflectType(&DefInt8)

	Int16    = newReflectType(DefInt16)
	Int16Ptr = newReflectType(&DefInt16)

	Int32    = newReflectType(DefInt32)
	Int32Ptr = newReflectType(&DefInt32)

	Int64    = newReflectType(DefInt64)
	Int64Ptr = newReflectType(&DefInt64)

	Float32    = newReflectType(DefFloat32)
	Float32Ptr = newReflectType(&DefFloat32)

	Float64    = newReflectType(DefFloat64)
	Float64Ptr = newReflectType(&DefFloat64)

	Bool    = newReflectType(DefBool)
	BoolPtr = newReflectType(&DefBool)

	Rune    = newReflectType(DefRune)
	RunePtr = newReflectType(&DefRune)

	Byte    = newReflectType(DefByte)
	BytePtr = newReflectType(&DefByte)
)

func newReflectType(v interface{}) reflect.Type {
	return reflect.TypeOf(v)
}

func newErrorReflectType() reflect.Type {
	type T struct {
		Error error
	}
	t := &T{}
	ot := reflect.TypeOf(t)
	field, _ := ot.Elem().FieldByName("Error")
	return field.Type
}

type Converter func(value string) (val interface{}, err error)

type ReflectType struct {
	Type      reflect.Type // 类型
	Converter Converter    // 转换器
	Name      string
}

// 类型映射
var types = make(map[reflect.Type]*ReflectType)

func RegisterType(v interface{}, converter Converter) {
	vType := reflect.TypeOf(v)
	if vType.Kind() == reflect.Ptr {
		types[vType] = &ReflectType{vType, converter, vType.String()}
		types[vType.Elem()] = &ReflectType{vType.Elem(), converter, vType.Elem().String()}
	} else {
		types[vType] = &ReflectType{vType, converter, vType.String()}
		vType = reflect.PtrTo(vType)
		types[vType] = &ReflectType{vType, converter, vType.String()}
	}
}

func init() {
	// 字符串
	RegisterType(DefString, func(value string) (val interface{}, err error) {
		val = value
		return
	})
	// int
	RegisterType(DefInt, func(value string) (val interface{}, err error) {
		if value == "" {
			return int(0), nil
		}
		val, err = xconv.ToInt(value)
		return
	})
	RegisterType(DefInt8, func(value string) (val interface{}, err error) {
		if value == "" {
			return int8(0), nil
		}
		val, err = xconv.ToInt8(value)
		return
	})
	RegisterType(DefInt16, func(value string) (val interface{}, err error) {
		if value == "" {
			return int16(0), nil
		}
		val, err = xconv.ToInt16(value)
		return
	})
	RegisterType(DefInt32, func(value string) (val interface{}, err error) {
		if value == "" {
			return int32(0), nil
		}
		val, err = xconv.ToInt32(value)
		return
	})
	RegisterType(DefInt64, func(value string) (val interface{}, err error) {
		if value == "" {
			return int64(0), nil
		}
		val, err = xconv.ToInt64(value)
		return
	})
	RegisterType(DefRune, func(value string) (val interface{}, err error) {
		if value == "" {
			return rune(0), nil
		}
		val, err = xconv.ToInt32(value)
		return
	})

	// int
	RegisterType(DefUint, func(value string) (val interface{}, err error) {
		if value == "" {
			return uint(0), nil
		}
		val, err = xconv.ToUint(value)
		return
	})
	RegisterType(DefUint8, func(value string) (val interface{}, err error) {
		if value == "" {
			return uint8(0), nil
		}
		val, err = xconv.ToUint8(value)
		return
	})
	RegisterType(DefUint16, func(value string) (val interface{}, err error) {
		if value == "" {
			return uint16(0), nil
		}
		val, err = xconv.ToUint16(value)
		return
	})
	RegisterType(DefUint32, func(value string) (val interface{}, err error) {
		if value == "" {
			return uint32(0), nil
		}
		val, err = xconv.ToUint32(value)
		return
	})
	RegisterType(DefUint64, func(value string) (val interface{}, err error) {
		if value == "" {
			return uint64(0), nil
		}
		val, err = xconv.ToUint64(value)
		return
	})
	RegisterType(DefByte, func(value string) (val interface{}, err error) {
		if value == "" {
			return byte(0), nil
		}
		val, err = xconv.ToUint8(value)
		return
	})

	RegisterType(DefBool, func(value string) (val interface{}, err error) {
		if value == "" {
			return false, nil
		}
		val, err = xconv.ToBool(value)
		return
	})

	RegisterType(DefFloat32, func(value string) (val interface{}, err error) {
		if value == "" {
			return float32(0), nil
		}
		val, err = xconv.ToFloat32(value)
		return
	})
	RegisterType(DefFloat64, func(value string) (val interface{}, err error) {
		if value == "" {
			return float64(0), nil
		}
		val, err = xconv.ToFloat64(value)
		return
	})
	RegisterType([]string{}, func(value string) (val interface{}, err error) {
		return toStringArray(value)
	})
	// bool
	RegisterType([]bool{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []int{}, nil
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]bool, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToBool(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})
	// int
	RegisterType([]int{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []int{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []int
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]int, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToInt(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})
	// uint
	RegisterType([]uint{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []uint{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []uint
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]uint, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToUint(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})

	// int8
	RegisterType([]int8{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []int8{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []int8
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]int8, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToInt8(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})
	// uint8
	RegisterType([]uint8{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []uint8{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []uint8
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]uint8, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToUint8(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})

	// int16
	RegisterType([]int16{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []int16{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []int16
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]int16, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToInt16(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})
	// uint16
	RegisterType([]uint16{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []uint16{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []uint16
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]uint16, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToUint16(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})

	// int32
	RegisterType([]int32{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []int32{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []int32
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]int32, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToInt32(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})

	// uint32
	RegisterType([]uint32{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []uint32{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []uint32
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]uint32, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToUint32(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})

	// int64
	RegisterType([]int64{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []int64{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []int64
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]int64, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToInt64(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})
	// uint64
	RegisterType([]uint64{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []uint64{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []uint64
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]uint64, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToUint64(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})
	// float32
	RegisterType([]float32{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []float32{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []float32
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}

		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]float32, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToFloat32(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})
	// float64
	RegisterType([]float64{}, func(value string) (val interface{}, err error) {
		if value == "" {
			return []float64{}, nil
		}
		value = xstr.Trim(value)
		if strings.HasPrefix(value, "[") {
			// 当成json 来解析
			var result []float64
			err := json.Unmarshal([]byte(value), &result)
			return result, err
		}
		sArr, err := toStringArray(value)
		if nil != err {
			return nil, err
		}
		arr := make([]float64, len(sArr))
		for i, sv := range sArr {
			v, err := xconv.ToFloat64(sv)
			if nil != err {
				return nil, err
			}
			arr[i] = v
		}
		return arr, nil
	})
}

func toStringArray(value string) ([]string, error) {
	if len(value) < 1 {
		return []string{}, nil
	}
	value = xstr.Trim(value)
	if strings.HasPrefix(value, "[") {
		// 当成json 来解析
		var result []string
		err := json.Unmarshal([]byte(value), &result)
		return result, err
	}
	// 非 json，那么按照英文逗号分隔
	arr := strings.Split(value, ",")
	for i, v := range arr {
		arr[i] = xstr.Trim(v)
	}
	return arr, nil
}
