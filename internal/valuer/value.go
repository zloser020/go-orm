package valuer

import (
	"database/sql"
	"reflect"
)

type Value interface {
	SetColumn(rows *sql.Rows) error
}

type Field struct {
	GoName string
	Typ    reflect.Type
	Offset uintptr
}

type Metadata interface {
	FieldByColumn(column string) (Field, bool)
}

type Creator func(model Metadata, entity any) Value
