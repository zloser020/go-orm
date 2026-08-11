package orm

import (
	"context"
	"orm/internal/errs"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	builder
	tableName string
	where     []Predicate
	expr      Expression

	db *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		builder: builder{
			sb: &strings.Builder{},
		},
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.model, err = s.db.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	sb := s.sb
	sb.WriteString("SELECT * FROM ")
	if s.tableName == "" {
		sb.WriteByte('`')
		sb.WriteString(s.model.tableName)
		sb.WriteByte('`')
	} else {
		sb.WriteString(s.tableName)
	}

	if len(s.where) > 0 {
		sb.WriteString(" WHERE ")
		if err := s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}

	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.tableName = table
	return s
}

// Where
// s.WHERE id IN (?, ?, ?), ids...
func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	db := s.db.db
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrNoRows
	}

	cs, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	tp := new(T)
	vals := make([]any, len(cs))
	valElem := make([]reflect.Value, len(cs))
	for _, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnkonwnColumn(c)
		}

		val := reflect.New(fd.typ)
		vals = append(vals, val.Interface())
		valElem = append(valElem, val.Elem())

	}

	err = rows.Scan(vals...)
	if err != nil {
		return nil, err
	}
	tpValueElem := reflect.ValueOf(tp).Elem()
	for i, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewErrUnkonwnColumn(c)
		}

		if fd.colName == c {
			tpValueElem.FieldByName(fd.goName).Set(valElem[i])
		}
	}
	return tp, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//q, err := s.Build()
	//if err != nil {
	//	return nil, err
	//}
	//db := s.db.db
	//rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	//if err != nil {
	//	return nil, err
	//}
	//for rows.Next() {
	//
	//}
	panic("unreachable")
}
