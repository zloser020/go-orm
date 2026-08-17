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

	for typ.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil, errors.New("value cannot be zero")
		}
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

func SetField(entity any, field string, newVal any) error {
	if entity == nil {
		return errors.New("value cannot be set")
	}
	val := reflect.ValueOf(entity)
	if val.Kind() != reflect.Pointer {
		return errors.New("value cannot be set")
	}
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return errors.New("value cannot be set")
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("value cannot be set")
	}
	fieldVal := val.FieldByName(field)
	if !fieldVal.IsValid() || !fieldVal.CanSet() {
		return errors.New("value cannot be set")
	}
	newValue := reflect.ValueOf(newVal)
	if !newValue.IsValid() {
		switch fieldVal.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			fieldVal.SetZero()
			return nil
		default:
			return errors.New("value type mismatch")
		}
	}
	if !newValue.Type().AssignableTo(fieldVal.Type()) {
		return errors.New("value type mismatch")
	}
	fieldVal.Set(newValue)
	return nil
}
