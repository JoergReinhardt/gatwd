package functions

import d "github.com/joergreinhardt/gatwd/data"

type (
	//// RECORD
	RecordField    func(...Callable) (Callable, string)
	RecordVal      func(...Callable) []RecordField
	RecordType     func(...Callable) RecordVal
	RecordTypeCons func(...Callable) RecordType
)

//// RECORD
///
//
func NewRecordType(defaults ...Callable) RecordType {
	var signature = createSignature(defaults...)
	var l = len(signature)
	var records []RecordField
	for i := 0; i < l; i++ {
		records = append(records, NewRecordField("", NewNone()))
	}
	records = createRecord(signature, records, defaults...)
	return func(args ...Callable) RecordVal {
		records = createRecord(signature, records, args...)
		return func(vals ...Callable) []RecordField {
			if len(vals) > 0 {
				records = applyRecord(signature, records, vals...)
			}
			return records
		}
	}
}

func applyRecord(
	signature []KeyPair,
	records []RecordField,
	args ...Callable,
) []RecordField {
	for _, arg := range args {
		for sigpos, sig := range signature {
			// apply record field
			if arg.TypeFnc().Match(Record | Element) {
				if recfield, ok := arg.(RecordField); ok {
					var val, key = recfield()
					var value = val.Call(records[sigpos])
					if sig.Value().(Paired).Left().TypeNat().Match(
						value.TypeNat(),
					) && sig.Value().(Paired).Left().TypeFnc().Match(
						value.TypeFnc(),
					) && sig.KeyStr() == recfield.KeyStr() {
						records[sigpos] = NewRecordField(key, value)
						break
					}
				}
			}
			// apply pair
			if arg.TypeFnc().Match(Key | Pair) {
				if pair, ok := arg.(Paired); ok {
					if pair.Left().TypeNat().Match(d.String) {
						if key, ok := pair.Left().Eval().(d.StrVal); ok {
							var value = pair.Right().Call(records[sigpos])
							if sig.Value().(Paired).Left().TypeNat().Match(
								value.TypeNat(),
							) && sig.Value().(Paired).Left().TypeFnc().Match(
								value.TypeFnc(),
							) && sig.KeyStr() == key.String() {
								records[sigpos] = NewRecordField(
									key.String(),
									value,
								)
								break
							}
						}
					}
				}
			}
		}
	}
	return records
}

func createRecord(
	signature []KeyPair,
	records []RecordField,
	args ...Callable,
) []RecordField {
	for _, arg := range args {
		for sigpos, sig := range signature {
			// apply record field
			if arg.TypeFnc().Match(Record | Element) {
				if recfield, ok := arg.(RecordField); ok {
					if sig.Value().(Paired).Left().TypeNat().Match(
						recfield.Value().TypeNat(),
					) && sig.Value().(Paired).Left().TypeFnc().Match(
						recfield.Value().TypeFnc(),
					) && sig.KeyStr() == recfield.KeyStr() {
						records[sigpos] = recfield
						break
					}
				}
			}
			// apply pair
			if arg.TypeFnc().Match(Key | Pair) {
				if pair, ok := arg.(Paired); ok {
					if pair.Left().TypeNat().Match(d.String) {
						if key, ok := pair.Left().Eval().(d.StrVal); ok {
							if sig.Value().(Paired).Left().TypeNat().Match(
								pair.Right().TypeNat(),
							) && sig.Value().(Paired).Right().TypeFnc().Match(
								pair.Right().TypeFnc(),
							) && sig.KeyStr() == key.String() {
								records[sigpos] = NewRecordField(
									key.String(),
									pair.Right(),
								)
								break
							}
						}
					}
				}
			}
		}
	}
	return records
}

func createSignature(
	args ...Callable,
) []KeyPair {
	var signature = make([]KeyPair, 0, len(args))
	for pos, arg := range args {
		// signature from record field argument
		if arg.TypeFnc().Match(Record | Element) {
			if field, ok := arg.(RecordField); ok {
				signature = append(signature, NewKeyPair(
					field.KeyStr(),
					NewPair(
						NewData(
							field.Value().TypeNat(),
						),
						NewData(
							field.Value().TypeFnc(),
						),
					)))
				continue
			}
		}
		// signature from pair argument
		if arg.TypeFnc().Match(Pair) {
			if pair, ok := arg.(Paired); ok {
				if pair.Left().TypeNat().Match(d.String) {
					if key, ok := pair.Left().Eval().(d.StrVal); ok {
						signature = append(signature, NewKeyPair(
							key.String(),
							NewPair(
								NewData(
									pair.Right().TypeNat(),
								),
								NewData(
									pair.Right().TypeFnc(),
								),
							)))
						continue
					}
				}
			}
		}
		// signature from alternating key/value arguments
		if arg.TypeNat().Match(d.String) {
			if key, ok := arg.Eval().(d.StrVal); ok {
				if len(args) > pos+1 {
					pos += 1
					arg = args[pos]
					signature = append(signature, NewKeyPair(
						key.String(),
						NewPair(
							NewData(
								arg.TypeNat(),
							),
							NewData(
								arg.TypeFnc(),
							),
						)))
					continue
				}
			}
			pos -= 1
		}
	}
	return signature
}

//// RECORD TYPE
func (t RecordType) Ident() Callable                { return t }
func (t RecordType) String() string                 { return t().String() }
func (t RecordType) TypeFnc() TyFnc                 { return Constructor | Record | t().TypeFnc() }
func (t RecordType) TypeNat() d.TyNat               { return d.Functor | t().TypeNat() }
func (v RecordType) Call(args ...Callable) Callable { return v(args...) }
func (v RecordType) Eval() d.Native                 { return v(natToFnc()...) }

//// RECORD VALUE
func (v RecordVal) Ident() Callable { return v }
func (v RecordVal) GetKey(key string) (RecordField, bool) {
	for _, field := range v() {
		if field.KeyStr() == key {
			return field, true
		}
	}
	return emptyRecordField(), false
}
func (v RecordVal) GetIdx(idx int) (RecordField, bool) {
	if idx < len(v()) {
		return v()[idx], true
	}
	return emptyRecordField(), false
}
func (v RecordVal) SetKey(key string, val Callable) (RecordVal, bool) {
	if _, ok := v.GetKey(key); ok {
		_ = v(NewRecordField(key, val))
		return v, true
	}
	return v, false
}
func (v RecordVal) SetIdx(idx int, val Callable) (RecordVal, bool) {
	if field, ok := v.GetIdx(idx); ok {
		_ = v(NewRecordField(field.KeyStr(), val))
		return v, true
	}
	return v, false
}
func (v RecordVal) Consume() (Callable, Consumeable) {
	var fields = v()
	if len(fields) > 0 {
		if len(fields) > 1 {
			var args = make([]Callable, 0, len(fields)-1)
			for _, field := range fields {
				args = append(args, field)
			}
			return fields[0],
				NewVector(args...)
		}
		return fields[0], v
	}
	return emptyRecordField(), v
}
func (v RecordVal) Head() Callable {
	if len(v()) > 0 {
		return v()[0]
	}
	return emptyRecordField()
}
func (v RecordVal) Tail() Consumeable {
	if len(v()) > 1 {
		var args = []Callable{}
		for _, field := range v()[1:] {
			args = append(args, field)
		}
		return NewVector(args...)
	}
	return NewNone()
}
func (v RecordVal) Call(args ...Callable) Callable {
	_ = v(args...)
	return v
}
func (v RecordVal) Eval() d.Native { return d.NewNil() }
func (v RecordVal) TypeNat() d.TyNat {
	var typ = d.Functor
	for _, field := range v() {
		typ = typ | field.TypeNat()
	}
	return typ
}
func (v RecordVal) TypeFnc() TyFnc {
	var typ = Record
	for _, field := range v() {
		typ = typ | field.TypeFnc()
	}
	return typ
}
func (v RecordVal) String() string {
	var l = len(v())
	var str = "("
	for pos, field := range v() {
		str = str + field.String()
		if pos < l-1 {
			str = str + ", "
		}
	}
	return str + ")"
}

//// RECORD FIELD
func emptyRecordField() RecordField {
	return func(...Callable) (Callable, string) { return NewNone(), "None" }
}
func NewRecordField(key string, val Callable) RecordField {
	return func(args ...Callable) (Callable, string) { return val, key }
}
func (a RecordField) String() string {
	return a.Key().String() + " :: " + a.Value().String()
}
func (a RecordField) Call(args ...Callable) Callable {
	return a.Right().Call(args...)
}
func (a RecordField) Eval() d.Native {
	return a.Value().Eval()
}
func (a RecordField) Both() (Callable, Callable) {
	var val, key = a()
	return NewData(d.StrVal(key)), val
}
func (a RecordField) Left() Callable {
	_, key := a()
	return NewData(d.StrVal(key))
}
func (a RecordField) Right() Callable {
	val, _ := a()
	return val
}
func (a RecordField) Empty() bool {
	if a.Left() == nil || (a.Right() == nil ||
		(!a.Right().TypeFnc().Flag().Match(None) ||
			!a.Right().TypeNat().Flag().Match(d.Nil))) {
		return true
	}
	return false
}
func (a RecordField) TypeNat() d.TyNat {
	return d.Functor |
		d.Pair |
		d.String |
		a.Value().TypeNat()
}
func (a RecordField) TypeFnc() TyFnc  { return Record | Element }
func (a RecordField) Key() Callable   { return a.Left() }
func (a RecordField) Value() Callable { return a.Right() }
func (a RecordField) Pair() Paired    { return NewPair(a.Both()) }
func (a RecordField) Pairs() []Paired { return []Paired{NewPair(a.Both())} }
func (a RecordField) KeyStr() string  { return a.Left().Eval().String() }
func (a RecordField) Ident() Callable { return a }
