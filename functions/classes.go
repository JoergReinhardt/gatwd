package functions

import ()

type (
	NumberVal func(...Callable) Numeral
	StringVal func(...Callable) Text
	RawVal    func(...Callable) Raw
)
