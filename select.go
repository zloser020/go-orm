package orm

import (
	"context"
	"strings"
)

type Selector[T any] struct {
	builder
	tableName string
	where     []Predicate
	expr      Expression
}

func (s *Selector[T]) Build() (*Query, error) {
	s.sb = &strings.Builder{}
	var err error
	s.model, err = parseModel(new(T))
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
	panic("not implemented")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("not implemented")
}
