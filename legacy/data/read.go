package data

import (
	"strconv"
	"time"
)

//// PARSE VALUES FROM STRING
///
// read methods expect a string lexeme, stripped of whitespace & parsabel as
// native value by strconv.
//
// BOOL
func (v StrVal) ReadBool() (bool, error) {
	var s, err = strconv.ParseBool(string(v))
	if err != nil {
		return false, err
	}
	return s, nil
}
func (v StrVal) ReadBoolVal() Native {
	var val, err = v.ReadBool()
	if err != nil {
		return NilVal{}
	}
	return BoolVal(val)
}

// NATURAL
func (v StrVal) ReadUint() (uint, error) {
	var u, err = strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}
func (v StrVal) ReadUintVal() Native {
	var val, err = v.ReadUint()
	if err != nil {
		return NilVal{}
	}
	return UintVal(val)
}

// INTEGER
func (v StrVal) ReadInt() (int, error) {
	var i, err = strconv.ParseInt(string(v), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}
func (v StrVal) ReadIntVal() Native {
	var val, err = v.ReadInt()
	if err != nil {
		return NilVal{}
	}
	return IntVal(val)
}

// FLOAT
func (v StrVal) ReadFloat() (float64, error) {
	var f, err = strconv.ParseFloat(string(v), 64)
	if err != nil {
		return 0.0, err
	}
	return float64(FltVal(f)), nil
}
func (v StrVal) ReadFloatVal() Native {
	var f, err = v.ReadFloat()
	if err != nil {
		return NilVal{}
	}
	return FltVal(f)
}

// DURATION
func (v StrVal) ReadDuration() (time.Duration, error) {
	var d, err = time.ParseDuration(v.String())
	if err != nil {
		return time.Duration(0), err
	}
	return d, nil
}
func (v StrVal) ReadDuraVal() Native {
	var dura, err = v.ReadDuration()
	if err != nil {
		return NilVal{}
	}
	return DuraVal(dura)
}

// TIME
func (v StrVal) ReadTime(layout string) (time.Time, error) {
	t, err := time.Parse(layout, v.String())
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}
func (v StrVal) ReadTimeVal(layout string) Native {
	var tim, err = v.ReadTime(layout)
	if err != nil {
		return NilVal{}
	}
	return TimeVal(tim)
}
