package utils

import (
	"errors"
	"github.com/gogf/gf/v2/util/gconv"
	"reflect"
)

// 是否在数组重
func InSlice[T comparable](a T, arr []T) bool {
	for _, v := range arr {
		if v == a {
			return true
		}
	}
	return false
}

func StructsToMap[T any](arr []T, field string) (map[string]T, error) {
	if len(arr) <= 0 {
		return nil, nil
	}
	if reflect.TypeOf(arr[0]).Kind() != reflect.Struct {
		return nil, errors.New("arr is not slice for struct")
	}
	m := make(map[string]T)
	for _, s := range arr {
		value := reflect.ValueOf(s).FieldByName(field)
		if !value.IsValid() {
			return nil, errors.New("struct not have field:" + field)
		}
		//if !value.Comparable() {
		//	return nil, errors.New("struct's field:" + field + " is not comparable")
		//}
		//为了兼容老版本的golang
		k := value.Type().Kind()
		if k == reflect.Array || k == reflect.Chan || k == reflect.Func || k == reflect.Interface ||
			k == reflect.Map || k == reflect.Slice || k == reflect.Struct || k == reflect.Pointer {
			return nil, errors.New("struct's field:" + field + " is not comparable")
		}
		m[gconv.String(value.Interface())] = s
	}
	return m, nil
}
