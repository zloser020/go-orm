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
	s.reset()
	var err error
	s.model, err = s.db.registry.Get(new(T))
	if err != nil {
		return nil, err
	}
	sb := s.sb
	sb.WriteString("SELECT * FROM ")
	if s.tableName == "" {
		sb.WriteByte('`')
		sb.WriteString(s.model.TableName)
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
	defer rows.Close()

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, err
		}
		return nil, ErrNoRows
	}

	tp := new(T)
	val := s.db.valuerCreator(s.model, tp)
	err = val.SetColumn(rows)
	if err != nil {
		return nil, err
	}
	return tp, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*T, 0)
	for rows.Next() {
		entity := new(T)
		val := s.db.valuerCreator(s.model, entity)
		if err = val.SetColumn(rows); err != nil {
			return nil, err
		}
		res = append(res, entity)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
