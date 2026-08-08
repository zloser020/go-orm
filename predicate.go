package orm

type op string

const (
	OpEq  op = "="
	opNot op = "NOT"
	opAnd op = "AND"
	opOr  op = "OR"
)

func (o op) String() string {
	return string(o)
}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

//func Eq(column string, arg any) Predicate {
//	return Predicate{
//		Column: column,
//		Op:     "=",
//		arg:    arg,
//	}
//}

type Column struct {
	name string
}

func C(name string) Column {
	return Column{
		name: name,
	}
}

// C("id").Eq(12)
func (c Column) Eq(arg any) Predicate {

	return Predicate{
		left:  c,
		op:    OpEq,
		right: value{val: arg},
	}
}

func (Column) expr() {}

// Not(C("id").Eq(12))
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// C("id").Eq(12).And(C.("name").Eq("George"))
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

// C("id").Eq(12).Or(C.("name").Eq("George"))
func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

func (Predicate) expr() {}

// Expression 标记接口
type Expression interface {
	expr()
}

type value struct {
	val any
}

func (value) expr() {}
