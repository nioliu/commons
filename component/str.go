package component

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
)

// ArrToStr 数组转字符串
func ArrToStr(arr interface{}, sep string) (str string, err error) {
	v := reflect.ValueOf(arr)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	if v.IsZero() {
		return "", nil
	}
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return "", errors.New("arr is not a slice type")
	}

	// 取样，判断类型
	sample := v.Index(0)
	if sample.Kind() == reflect.Pointer {
		if sample.IsNil() {
			return "", nil
		}
		sample = sample.Elem()
	}

	res := ""
	switch k := sample.Kind(); k {
	case reflect.Struct:
		for i := 0; i < v.Len(); i++ {
			marshal, err := json.Marshal(v.Index(i).Interface())
			if err != nil {
				return "", err
			}
			res += string(marshal) + sep
			if i == v.Len()-1 {
				res += string(marshal)
			}
		}
	case reflect.String:
		for i := 0; i < v.Len(); i++ {
			res += v.Index(i).String() + sep
			if i == v.Len()-1 {
				res += v.Index(i).String()
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for i := 0; i < v.Len(); i++ {
			if i == v.Len()-1 {
				res += strconv.Itoa(int(v.Index(i).Int()))
			} else {
				res += strconv.Itoa(int(v.Index(i).Int())) + sep
			}
		}
	}

	return res, nil
}

func StrArrToStr(arr []string, sep string) (string, error) {
	if arr == nil {
		return "", nil
	}
	if sep == "" {
		return "", errors.New("lack sep parameter")
	}
	var res = ""
	for i, s := range arr {
		res += s + sep
		if i == len(arr)-1 {
			res += s
		}
	}
	return res, nil
}
