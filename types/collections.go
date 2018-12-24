package types

type ConCellFn func() EmptyCellFn

func (c ConCellFn) Type() flag   { return Cell.Type() }
func (c ConCellFn) Name() strVal { return "List" }
func (c ConCellFn) Eval(...Data) Data { // :: a,a,a,a -> [a,a,a,a]
	var dat Data
	// TODO: can only be implemented, once everything else is set up
	return dat
}

type EmptyCellFn func(...Data) CellFn

func (c EmptyCellFn) Type() flag   { return Cell.Type() }
func (c EmptyCellFn) Name() strVal { return "[]" }
func (c EmptyCellFn) Eval(d ...Data) Data { // :: "Type" -> CellFn // aka lookup in cf-map
	//TODO: actual lookup... for now, just create
	return c(d...)
}

///// noooo, evan better shift everything one up! what does the empty cell
// constructor needs his own type for!?
//
// one more thing to be inserted here...  keep instances of typed cell function
// constructors around for re-use...  produce cell-function-instance as
// anonymously typed closure with most explicit types (does not need to be
// stored anywhere), single instance that get's passed around is immutable.
////

type CellFn func(...Data) func(Data) (Data, CellFn)

func (c CellFn) Type() flag     { return Cell.Type() }
func (c CellFn) String() strVal { return "[" + c.Type().String() + "]" }
func (c CellFn) Name() strVal   { return "x:xs" }
func (c CellFn) Eval(...Data) Data { // :: [x₁] -> x₁, [x₂]
	var dat Data
	return dat
}

var cellCons map[strVal]CellFn
