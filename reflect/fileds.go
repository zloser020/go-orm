package reflect

import (
	"errors"
	"reflect"
)

func IterateFields(entity any) (map[string]any, error) {
	if entity == nil {
		return nil, errors.New("entity cannot be nil")
	}

	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	if val.IsZero() {
		return nil, errors.New("value cannot be zero")
	}

	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.New("entity must be a struct")
	}
	numFields := typ.NumField()
	res := make(map[string]any, numFields)
	for i := 0; i < numFields; i++ {
		// 字段类型
		fieldType := typ.Field(i)
		// 字段值
		fieldValue := val.Field(i)
		if fieldType.IsExported() {
			res[fieldType.Name] = fieldValue.Interface()
		} else {
			res[fieldType.Name] = reflect.Zero(fieldType.Type).Interface()
		}

	}
	return res, nil
}
