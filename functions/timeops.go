package functions

import (
	"fmt"
	"math/big"
)

// This file is the operator matrix of the three time types (Duration, Time, Date). The
// rule behind it is the one for points and vectors (an affine space): a Duration is a
// vector, Time and Date are points.
//
//	Duration ± Duration            Duration
//	Duration * n, n * Duration     Duration   (n an Int, Ratio or Float; truncated toward zero)
//	Duration / n                   Duration   (n non-zero)
//	Duration / Duration            the exact ratio: a Ratio, demoted to an Int when whole
//	Duration // Duration           the floor quotient, an Int
//	Duration % Duration            the remainder, a Duration with the sign of the divisor
//	-Duration, abs                 Duration
//	Time + Duration, Time - Duration   Time      (in the same display zone)
//	Time - Time                        Duration  (an error beyond about ±292 years)
//	==, !=, <, <=, >, >=           within one type only (a Time by instant, not by zone)
//
// Everything else is an error that names both types: there is no implicit conversion
// between a Duration and an Int, so `5s + 1` and `5s * 5s` are mistakes the language
// reports instead of guessing a unit.

// temporalKind names the time type of v ("Duration", "Time" or "Date") or is empty.
func temporalKind(v any) string {
	switch v.(type) {
	case Duration:
		return "Duration"
	case Time:
		return "Time"
	}
	return ""
}

// IsTemporal reports whether v is a Duration, a Time or a Date.
func IsTemporal(v any) bool { return temporalKind(v) != "" }

// temporalBinaryOp evaluates a binary operator when at least one operand is a Duration,
// Time or Date. handled is false when neither is (the caller carries on) and for the
// equality operators whose operands do not mix time types (the caller's structural
// equality answers those).
func temporalBinaryOp(op OpKind, l, r any) (res any, handled bool, err error) {
	lk, rk := temporalKind(l), temporalKind(r)
	if lk == "" && rk == "" {
		return nil, false, nil
	}
	switch op {
	case OpConcat, OpAnd, OpOr:
		return nil, false, nil // the generic code reports these
	case OpEq, OpNeq:
		if lk == "" || rk == "" {
			return nil, false, nil // a time value never equals a value of another kind
		}
		if lk != rk {
			return nil, true, mixedTemporal(op, lk, rk, "")
		}
		c, _ := temporalCompare(l, r)
		return (c == 0) == (op == OpEq), true, nil
	case OpLt, OpLte, OpGt, OpGte:
		if lk != rk {
			return nil, true, mixedTemporal(op, kindOrName(l, lk), kindOrName(r, rk), "")
		}
		c, err := temporalCompare(l, r)
		if err != nil {
			return nil, true, err
		}
		return orderingHolds(op, c), true, nil
	}
	if lk == "Time" || rk == "Time" {
		res, err = timeArith(op, l, r)
		return res, true, err
	}
	if lk == "Duration" || rk == "Duration" {
		res, err = durationArith(op, l, r)
		return res, true, err
	}
	return nil, true, mixedTemporal(op, kindOrName(l, lk), kindOrName(r, rk), "")
}

func kindOrName(v any, kind string) string {
	if kind != "" {
		return kind
	}
	return KindName(v)
}

func mixedTemporal(op OpKind, l, r, hint string) error {
	if hint != "" {
		hint = "; " + hint
	}
	return fmt.Errorf("operator '%s' is not defined for %s and %s%s", op, l, r, hint)
}

// temporalCompare orders two values of the same time type.
func temporalCompare(a, b any) (int, error) {
	if x, ok := a.(Duration); ok {
		if y, ok := b.(Duration); ok {
			switch {
			case x < y:
				return -1, nil
			case x > y:
				return 1, nil
			}
			return 0, nil
		}
	}
	if x, ok := a.(Time); ok {
		if y, ok := b.(Time); ok {
			return x.Compare(y), nil
		}
	}
	return 0, fmt.Errorf("cannot compare %s and %s", KindName(a), KindName(b))
}

// durationArith is the Duration row of the matrix (see the table at the top of the file).
func durationArith(op OpKind, l, r any) (any, error) {
	ld, lIs := l.(Duration)
	rd, rIs := r.(Duration)
	bad := func(hint string) (any, error) {
		return nil, mixedTemporal(op, KindName(l), KindName(r), hint)
	}
	switch {
	case lIs && rIs:
		switch op {
		case OpAdd, OpSub:
			a, b := int64(ld), int64(rd)
			if op == OpSub {
				if b == minInt64 {
					return nil, durationOverflow(op, ld, rd)
				}
				b = -b
			}
			c := a + b
			if (c > a) != (b > 0) {
				return nil, durationOverflow(op, ld, rd)
			}
			return Duration(c), nil
		case OpDiv, OpIDiv, OpMod:
			if rd == 0 {
				return nil, divZero(op)
			}
			q := new(big.Rat).SetFrac(big.NewInt(int64(ld)), big.NewInt(int64(rd)))
			switch op {
			case OpDiv:
				return CanonRat(q), nil
			case OpIDiv:
				return CanonInt(new(big.Int).Div(q.Num(), q.Denom())), nil // denominator > 0: floor
			}
			fl := new(big.Int).Div(q.Num(), q.Denom())
			rem := new(big.Int).Sub(big.NewInt(int64(ld)), fl.Mul(fl, big.NewInt(int64(rd))))
			return Duration(rem.Int64()), nil // |rem| < |rd|, so it fits
		case OpMul:
			return bad("multiply a Duration by a number, not by another Duration")
		}
		return bad("")
	case lIs: // Duration op number
		switch op {
		case OpMul, OpDiv:
			f, err := scaleFactor(r)
			if err != nil {
				return nil, err
			}
			if op == OpDiv {
				if f.Sign() == 0 {
					return nil, divZero(op)
				}
				f = new(big.Rat).Inv(f)
			}
			return durationFromRat(new(big.Rat).Mul(new(big.Rat).SetInt64(int64(ld)), f), "Duration "+string(op))
		}
		return bad("")
	default: // number op Duration
		switch op {
		case OpMul:
			f, err := scaleFactor(l)
			if err != nil {
				return nil, err
			}
			return durationFromRat(new(big.Rat).Mul(new(big.Rat).SetInt64(int64(rd)), f), "Duration *")
		case OpDiv:
			return bad("divide a Duration by a number, or by another Duration; a number cannot be divided by a length of time")
		}
		return bad("")
	}
}

const minInt64 = -1 << 63

func durationOverflow(op OpKind, a, b Duration) error {
	return fmt.Errorf("Duration overflow: %s %s %s is outside about ±292 years", a, op, b)
}

// timeArith is the Time row of the matrix: a point plus or minus a vector is a point, a
// point minus a point is a vector, and every other combination is an error.
func timeArith(op OpKind, l, r any) (any, error) {
	lt, lIs := l.(Time)
	rt, rIs := r.(Time)
	switch {
	case lIs && rIs:
		if op == OpSub {
			return timeSub(lt, rt)
		}
		if op == OpAdd {
			return nil, mixedTemporal(op, "Time", "Time", "two instants cannot be added; subtract them to get a Duration")
		}
	case lIs:
		if d, ok := r.(Duration); ok {
			switch op {
			case OpAdd:
				return timeAdd(lt, d)
			case OpSub:
				neg, err := Negate(d)
				if err != nil {
					return nil, err
				}
				return timeAdd(lt, neg.(Duration))
			}
		}
	case rIs:
		if _, ok := l.(Duration); ok && op == OpAdd {
			return nil, mixedTemporal(op, "Duration", "Time", "write the Time first: t + d")
		}
		if _, ok := l.(Duration); ok && op == OpSub {
			return nil, mixedTemporal(op, "Duration", "Time", "a Duration minus an instant has no meaning; subtract the other way round")
		}
	}
	return nil, mixedTemporal(op, KindName(l), KindName(r), "")
}
