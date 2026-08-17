package unsafe

import "reflect"

func PrintFieldOffset(entity any) {
	typ := reflect.TypeOf(entity)
	if typ == nil {
		return
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		println(field.Offset)
	}
}
