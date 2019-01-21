package lang

import d "github.com/JoergReinhardt/godeep/data"

//////////////////////////
// input item data interface
type Item interface {
	ItemType() d.BitFlag
	String() string
	Value() d.Data
}
