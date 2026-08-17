package unsafe

import (
	"fmt"
	"reflect"
	"unsafe"
)

type UnsafeAccessor struct {
	fields  map[string]FieldMeta
	address unsafe.Pointer
	err     error
}

func NewUnsafeAccessor(entity any) *UnsafeAccessor {
	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	if typ == nil || typ.Kind() != reflect.Ptr || val.IsNil() || typ.Elem().Kind() != reflect.Struct {
		return &UnsafeAccessor{err: fmt.Errorf("entity must be a non-nil struct pointer")}
	}
	typ = typ.Elem()
	numField := typ.NumField()
	fields := make(map[string]FieldMeta, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fields[fd.Name] = FieldMeta{
			Offset: fd.Offset,
			typ:    fd.Type,
		}
	}
	return &UnsafeAccessor{
		fields:  fields,
		address: val.UnsafePointer(),
	}
}

func (a *UnsafeAccessor) Field(field string) (any, error) {
	if a.err != nil {
		return nil, a.err
	}
	fd, ok := a.fields[field]
	if !ok {
		return nil, fmt.Errorf("field %s not found", field)
	}
	// 字段起始地址
	fdAddr := unsafe.Add(a.address, fd.Offset)
	// 不知道确切类型
	return reflect.NewAt(fd.typ, fdAddr).Elem().Interface(), nil
}

func (a *UnsafeAccessor) SetField(field string, value any) error {
	if a.err != nil {
		return a.err
	}
	fd, ok := a.fields[field]
	if !ok {
		return fmt.Errorf("field %s not found", field)
	}
	fdAddr := unsafe.Add(a.address, fd.Offset)

	fieldValue := reflect.NewAt(fd.typ, fdAddr).Elem()
	newValue := reflect.ValueOf(value)
	if !newValue.IsValid() {
		switch fieldValue.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			fieldValue.SetZero()
			return nil
		default:
			return fmt.Errorf("value type mismatch: cannot assign nil to %s", fd.typ)
		}
	}
	if !newValue.Type().AssignableTo(fd.typ) {
		return fmt.Errorf("value type mismatch: cannot assign %s to %s", newValue.Type(), fd.typ)
	}
	fieldValue.Set(newValue)
	return nil
}

type FieldMeta struct {
	Offset uintptr
	typ    reflect.Type
}
