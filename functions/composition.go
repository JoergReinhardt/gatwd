package functions

type (
	SequenceVal func(...Expression) (Expression, SequenceVal)
)

func NewSequence(coll Consumeable) SequenceVal {
	return func(args ...Expression) (Expression, SequenceVal) {
		if len(args) > 0 {
			return NewSequence(coll.Append(args...))()
		}
		var expr, coll = coll.Consume()
		return expr, NewSequence(coll)
	}
}

//func (s SequenceVal) String() string { return }
