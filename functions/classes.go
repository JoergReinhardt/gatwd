package functions

import ()

type (
	NumberVal func(...Callable) Numeral
	StringVal func(...Callable) Text
	RawBytes  func(...Callable) Raw
)
