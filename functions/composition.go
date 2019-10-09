package functions

type (
	SequenceVal func(...Expression) (Expression, Sequential)
)

func NewSequence(coll Sequential) SequenceVal {
	return func(args ...Expression) (Expression, Sequential) {
		if len(args) > 0 {
			return coll.Cons(args...).Consume()
		}
		return coll.Consume()
	}
}

func (s SequenceVal) TypeFnc() TyFnc      { return Sequence }
func (s SequenceVal) Type() TyPattern     { return s.Tail().Type() }
func (s SequenceVal) TypeElem() TyPattern { return s.Head().Type() }
func (s SequenceVal) Cons(elems ...Expression) Sequential {
	return SequenceVal(func(args ...Expression) (Expression, Sequential) {
		if len(args) > 0 {
			return s.Cons(elems...).Cons(args...).Consume()
		}
		return s.Cons(elems...).Consume()
	})
}
func (s SequenceVal) Append(elems ...Expression) Sequential {
	return SequenceVal(func(args ...Expression) (Expression, Sequential) {
		if len(args) > 0 {
			return s.Append(elems...).Cons(args...).Consume()
		}
		return s.Append(elems...).Consume()
	})
}
func (s SequenceVal) Call(args ...Expression) Expression {
	var head Expression
	if len(args) > 0 {
		head, _ = s(args...)
		return head
	}
	head, _ = s()
	return head
}
func (s SequenceVal) Consume() (Expression, Sequential) { return s() }
func (s SequenceVal) Head() Expression {
	var expr, _ = s()
	return expr
}
func (s SequenceVal) Tail() Sequential {
	var _, coll = s()
	return coll
}
func (s SequenceVal) String() string {
	var head, tail = s()
	return tail.Cons(head).String()
}
