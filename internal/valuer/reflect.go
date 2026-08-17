package valuer

import (
	"database/sql"
	"orm/internal/errs"
	"reflect"
)

type reflectValue struct {
	model Metadata
	// T的指针
	val any
}

var _ Creator = NewReflectValue

func NewReflectValue(model Metadata, val any) Value {
	return &reflectValue{model: model, val: val}
}

func (r reflectValue) SetColumn(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	vals := make([]any, len(cs))
	valElem := make([]reflect.Value, len(cs))
	for i, c := range cs {
		fd, ok := r.model.FieldByColumn(c)
		if !ok {
			return errs.NewErrUnkonwnColumn(c)
		}

		val := reflect.New(fd.Typ)
		vals[i] = val.Interface()
		valElem[i] = val.Elem()
	}

	err = rows.Scan(vals...)
	if err != nil {
		return err
	}
	tpValue := reflect.ValueOf(r.val)
	if tpValue.Kind() != reflect.Ptr || tpValue.IsNil() || tpValue.Elem().Kind() != reflect.Struct {
		return errs.ErrPointerOnly
	}
	tpValueElem := tpValue.Elem()
	for i, c := range cs {
		fd, ok := r.model.FieldByColumn(c)
		if !ok {
			return errs.NewErrUnkonwnColumn(c)
		}

		field := tpValueElem.FieldByName(fd.GoName)
		if !field.IsValid() || !field.CanSet() {
			return errs.NewErrUnkonwnField(fd.GoName)
		}
		field.Set(valElem[i])
	}
	return nil
}
