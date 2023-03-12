package component

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func TransType(raw interface{}, aimKind reflect.Kind) (interface{}, error) {
	value := reflect.ValueOf(raw)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if value.IsZero() {
		return nil, nil
	}

	switch kind := value.Kind(); kind {
	case reflect.String:
		return transStrType(value.String(), aimKind)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		intRaw := int(value.Int())
		strRaw := strconv.Itoa(intRaw)
		return transStrType(strRaw, aimKind)
	case reflect.Float32, reflect.Float64:
		valueStr := fmt.Sprintf("%f", value.Float())
		return transStrType(valueStr, aimKind)
	default:
		return nil, errors.New("unsupported raw kind")
	}
}

func transStrType(raw string, aimKind reflect.Kind) (interface{}, error) {
	var res interface{}
	var err error

	switch aimKind {
	case reflect.String:
		res = raw
	case reflect.Int:
		res, err = transStringToInt(raw)
	case reflect.Int8:
		res, err = transStringToInt(raw)
		res = int8(res.(int))
	case reflect.Int16:
		res, err = transStringToInt(raw)
		res = int16(res.(int))
	case reflect.Int32:
		res, err = transStringToInt(raw)
		res = int32(res.(int))
	case reflect.Int64:
		res, err = transStringToInt(raw)
		res = int64(res.(int))
	case reflect.Float32:
		res, err = transStringToFloat64(raw)
	case reflect.Float64:
		res, err = transStringToFloat32(raw)
	}

	return res, err
}

func transStringToInt(raw string) (int, error) {
	return strconv.Atoi(raw)
}

func transStringToFloat64(raw string) (float64, error) {
	return strconv.ParseFloat(raw, 64)
}

func transStringToFloat32(raw string) (float32, error) {
	f, err := strconv.ParseFloat(raw, 32)
	return float32(f), err
}
