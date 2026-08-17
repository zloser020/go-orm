package orm

import (
	"strings"
)

type Deleter[T any] struct {
	builder
	tableName string
	where     []Predicate
	expr      Expression

	db *DB
}

func NewDeleter[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		builder: builder{
			sb: &strings.Builder{},
		},
		db: db,
	}
}

func (d *Deleter[T]) Build() (*Query, error) {
	d.reset()
	var err error
	d.model, err = d.db.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	sb := d.sb
	sb.WriteString("DELETE FROM ")
	if d.tableName == "" {
		sb.WriteByte('`')
		sb.WriteString(d.model.TableName)
		sb.WriteByte('`')
	} else {
		sb.WriteByte('`')
		sb.WriteString(d.tableName)
		sb.WriteByte('`')
	}
	if len(d.where) > 0 {
		sb.WriteString(" WHERE ")
		if err := d.buildPredicates(d.where); err != nil {
			return nil, err
		}
	}
	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deleter[T]) From(tableName string) *Deleter[T] {
	d.tableName = tableName
	return d
}

func (d *Deleter[T]) Where(w ...Predicate) *Deleter[T] {
	d.where = w
	return d
}
