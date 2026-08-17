package valuer

import (
	"database/sql"
	"orm/internal/errs"
	"reflect"
	"runtime"
	"unsafe"
)

type UnsafeValue struct {
	model Metadata
	val   any
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(model Metadata, val any) Value {
	return &UnsafeValue{model: model, val: val}
}

func (u UnsafeValue) SetColumn(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	entity := reflect.ValueOf(u.val)
	if entity.Kind() != reflect.Ptr || entity.IsNil() || entity.Elem().Kind() != reflect.Struct {
		return errs.ErrPointerOnly
	}

	vals := make([]any, len(cs))
	address := entity.UnsafePointer()
	for i, c := range cs {
		fd, ok := u.model.FieldByColumn(c)
		if !ok {
			return errs.NewErrUnkonwnColumn(c)
		}

		fdAddress := unsafe.Add(address, fd.Offset)
		val := reflect.NewAt(fd.Typ, fdAddress)
		vals[i] = val.Interface()
	}
	err = rows.Scan(vals...)
	runtime.KeepAlive(u.val)
	return err
}
