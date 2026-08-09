package orm

import (
	"context"
	"fmt"
	"orm/internal/errs"
	"strings"
)

type Selector[T any] struct {
	tableName string
	model     *model
	where     []Predicate
	sb        *strings.Builder
	args      []any
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
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}

		if err := s.BuildExpression(p); err != nil {
			return nil, err
		}
	}

	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) BuildExpression(expr Expression) error {
	switch expr := expr.(type) {
	case nil:
		return nil
	case Predicate:
		_, ok := expr.left.(Predicate)
		if ok {
			s.sb.WriteString("(")
		}
		if err := s.BuildExpression(expr.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteString(")")
		}

		s.sb.WriteByte(' ')
		s.sb.WriteString(expr.op.String())
		s.sb.WriteByte(' ')

		_, ok = expr.right.(Predicate)
		if ok {
			s.sb.WriteString("(")
		}
		if err := s.BuildExpression(expr.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteString(")")
		}

	case Column:
		// 字段校验
		fd, ok := s.model.fields[expr.name]
		if !ok {
			return errs.NewErrUnkonwnField(expr.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.colName)
		s.sb.WriteByte('`')
	case value:
		s.sb.WriteByte('?')
		s.addArgs(expr.val)
	default:
		return fmt.Errorf("orm: invalid expression type: %T", expr)
	}
	return nil
}

func (s *Selector[T]) addArgs(val any) {
	if s.args == nil {
		s.args = make([]any, 0, 4)
	}
	s.args = append(s.args, val)
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
