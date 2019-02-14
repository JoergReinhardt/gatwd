/*
  HEAP OBJECT CONSTRUCTORS

    this file contains implementations of constructors for heap objects. they
    use the constructors from functions/constructors.go &
    functions/functions.go as closure to be evaluated to form the constructors
    for inbuildt static types of closures. describes and parametrizes them by
    creating appropriate info tables and defining and instanciating types to
    hold info and data associated with the particular kind of heap-object they
    construct.
*/
package run

import (
	d "github.com/JoergReinhardt/gatwd/data"
	f "github.com/JoergReinhardt/gatwd/functions"
)

func allocateAtomicConstant(prime d.Primary) *Object {

	var closure f.Value

	switch prime.TypePrime() {
	case d.Flag:
		var value = prime.(d.BitFlag)
		closure = f.NewPrimaryConstatnt(value)
	case d.Nil:
		var value = prime.(d.NilVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Bool:
		var value = prime.(d.BoolVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Int:
		var value = prime.(d.IntVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Int8:
		var value = prime.(d.Int8Val)
		closure = f.NewPrimaryConstatnt(value)
	case d.Int16:
		var value = prime.(d.Int16Val)
		closure = f.NewPrimaryConstatnt(value)
	case d.Int32:
		var value = prime.(d.Int32Val)
		closure = f.NewPrimaryConstatnt(value)
	case d.Uint:
		var value = prime.(d.UintVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Uint8:
		var value = prime.(d.Uint8Val)
		closure = f.NewPrimaryConstatnt(value)
	case d.Uint16:
		var value = prime.(d.Uint16Val)
		closure = f.NewPrimaryConstatnt(value)
	case d.Uint32:
		var value = prime.(d.Uint32Val)
		closure = f.NewPrimaryConstatnt(value)
	case d.Float:
		var value = prime.(d.FltVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Flt32:
		var value = prime.(d.Flt32Val)
		closure = f.NewPrimaryConstatnt(value)
	case d.Imag:
		var value = prime.(d.ImagVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Imag64:
		var value = prime.(d.Imag64Val)
		closure = f.NewPrimaryConstatnt(value)
	case d.Byte:
		var value = prime.(d.ByteVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Rune:
		var value = prime.(d.RuneVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Bytes:
		var value = prime.(d.BytesVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.String:
		var value = prime.(d.StrVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.BigInt:
		var value = prime.(d.BigIntVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.BigFlt:
		var value = prime.(d.BigFltVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Ratio:
		var value = prime.(d.RatioVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Time:
		var value = prime.(d.TimeVal)
		closure = f.NewPrimaryConstatnt(value)
	case d.Duration:
		var value = prime.(d.DuraVal)
		closure = f.NewPrimaryConstatnt(value)
	}
	var object = allocateObject()
	(*object).Info.Length = Length(1)
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = Propertys(Data | Atomic)
	(*object).Otype = DataConstructor
	(*object).Value = closure
	return object
}
func allocateVectorConstant(prime d.Primary) *Object {

	var closure f.Value

	switch {
	case prime.TypePrime().Flag().Match(d.Vector | d.Flag):
		var value = prime.(d.BitFlag)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Nil):
		var value = prime.(d.NilVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Bool):
		var value = prime.(d.BoolVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Int):
		var value = prime.(d.IntVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Int8):
		var value = prime.(d.Int8Vec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Int16):
		var value = prime.(d.Int16Vec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Int32):
		var value = prime.(d.Int32Vec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Uint):
		var value = prime.(d.UintVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Uint8):
		var value = prime.(d.Uint8Vec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Uint16):
		var value = prime.(d.Uint16Vec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Uint32):
		var value = prime.(d.Uint32Vec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Float):
		var value = prime.(d.FltVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Flt32):
		var value = prime.(d.Flt32Vec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Imag):
		var value = prime.(d.ImagVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Imag64):
		var value = prime.(d.Imag64Vec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Byte):
		var value = prime.(d.ByteVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Rune):
		var value = prime.(d.RuneVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Bytes):
		var value = prime.(d.BytesVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.String):
		var value = prime.(d.StrVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.BigInt):
		var value = prime.(d.BigIntVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.BigFlt):
		var value = prime.(d.BigFltVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Ratio):
		var value = prime.(d.RatioVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Time):
		var value = prime.(d.TimeVec)
		closure = f.NewPrimaryConstatnt(value)
	case prime.TypePrime().Flag().Match(d.Vector | d.Duration):
		var value = prime.(d.DuraVec)
		closure = f.NewPrimaryConstatnt(value)
	}
	var object = allocateObject()
	(*object).Info.Length = Length(closure.(d.Sliceable).Len())
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = Propertys(Data)
	(*object).Otype = DataConstructor
	(*object).Value = closure
	return object
}

func allocatePrimarySet(pairs ...d.PairVal) *Object {
	// allocate flag to safe and compare accessors types
	var accflag d.BitFlag
	var set d.Mapped
	for i, pair := range pairs {
		// take accessor type of first
		// pair to decide on type of
		// slice accessors for all
		// succeeding pairs
		if i == 0 {
			// safe accessor flag
			// to compare against
			// succeeding pairs
			accflag = pair.Left().TypePrime().Flag()
			// allocate appropriate type of set
			switch accflag {
			case d.String.Flag():
				set = d.SetString{}
			case d.Flag.Flag():
				set = d.SetFlag{}
			case d.Float.Flag():
				set = d.SetFloat{}
			case d.Uint.Flag():
				set = d.SetInt{}
			case d.Int.Flag():
				set = d.SetUint{}
			}
		}
		// if this pairs accessor flag
		// happens to match the sets
		// accessor type
		if accflag.Match(pair.Left().TypePrime()) {
			// use interface method to set a new member
			set.Set(pair.Left(), pair.Right())

		}
	}
	var object = allocateObject()
	(*object).Info.Length = Length(len(set.(d.Mapped).Keys()))
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = Propertys(Data)
	(*object).Otype = DataConstructor
	(*object).Value = f.NewNaryFnc(func(...f.Value) f.Value {
		return f.NewPrimaryConstatnt(set)
	})
	return object
}
func allocatePrimaryDataSlice(data ...d.Primary) *Object {
	// make it a data slice
	var slice = d.NewSlice(data...)
	// generate a constant closure enclosing slice
	// allocate heap object
	var object = allocateObject()
	(*object).Info.Length = Length(slice.Len())
	(*object).Info.Arity = Arity(0)
	(*object).Info.Propertys = Propertys(Data)
	(*object).Otype = DataConstructor
	(*object).Value = f.NewNaryFnc(func(...f.Value) f.Value {
		return f.NewPrimaryConstatnt(slice)
	})
	return object
}
func allocatePrimaryData(data ...d.Primary) *Object {
	if len(data) == 1 {
		// when dealing single instance of a primary, return atomic primary instance
		return allocateAtomicConstant(data[0])
	}
	// more than a single instance → try to allocate a collection
	if len(data) > 1 {
		var flag = d.BitFlag(0)
		// concatenate flags
		for _, prim := range data {
			flag = flag | prim.TypePrime().Flag()
		}
		// if all instances have the same type, try to allocate slice
		// of natives
		if flag.Count() == 1 {
			// if it's an array of instances of the same type
			return allocateVectorConstant(
				d.ConNativeSlice(flag, data...),
			)
		}
		// if the set bit is set, try to allocate set
		if flag.Match(d.Pair) {
			var pairs = []d.PairVal{}
			for _, dat := range data {
				if pair, ok := dat.(d.PairVal); ok {
					pairs = append(pairs, pair)
				}
			}
			return allocatePrimarySet(pairs...)
		}
		// since nothing else seems to fit this collection, allocate a
		// slice of mixed primary type values (ein kessel buntes)
		return allocatePrimaryDataSlice(data...)
	}
	// arguments length is zero. return a nil value
	return allocateAtomicConstant(d.NilVal{})
}
