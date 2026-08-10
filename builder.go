package orm

import (
	"fmt"
	"orm/internal/errs"
	"strings"
)

type builder struct {
	sb    *strings.Builder
	args  []any
	model *Model
}

func (b *builder) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return b.BuildExpression(p)
}

func (b *builder) BuildExpression(expr Expression) error {
	switch expr := expr.(type) {
	case nil:
		return nil
	case Predicate:
		_, ok := expr.left.(Predicate)
		if ok {
			b.sb.WriteString("(")
		}
		if err := b.BuildExpression(expr.left); err != nil {
			return err
		}
		if ok {
			b.sb.WriteString(")")
		}

		b.sb.WriteByte(' ')
		b.sb.WriteString(expr.op.String())
		b.sb.WriteByte(' ')

		_, ok = expr.right.(Predicate)
		if ok {
			b.sb.WriteString("(")
		}
		if err := b.BuildExpression(expr.right); err != nil {
			return err
		}
		if ok {
			b.sb.WriteString(")")
		}

	case Column:
		// 字段校验
		fd, ok := b.model.fields[expr.name]
		if !ok {
			return errs.NewErrUnkonwnField(expr.name)
		}
		b.sb.WriteByte('`')
		b.sb.WriteString(fd.colName)
		b.sb.WriteByte('`')
	case value:
		b.sb.WriteByte('?')
		b.addArgs(expr.val)
	default:
		return fmt.Errorf("orm: invalid expression type: %T", expr)
	}
	return nil
}

func (b *builder) addArgs(val any) {
	if b.args == nil {
		b.args = make([]any, 0, 4)
	}
	b.args = append(b.args, val)
}
