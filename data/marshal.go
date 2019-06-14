package data

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"time"
)

// provide serialization for all native types.
func (v BitFlag) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v FlagSlice) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(v)))
	err := binary.Write(buf, binary.LittleEndian, v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (v PairVal) MarshalBinary() ([]byte, error) {
	buf0, err0 := v.Left().(BinaryMarshaler).MarshalBinary()
	if err0 != nil {
		return nil, err0
	}
	buf1, err1 := v.Right().(BinaryMarshaler).MarshalBinary()
	if err1 != nil {
		return nil, err1
	}
	return append(buf0, buf1...), nil
}

func (v NilVal) MarshalBinary() ([]byte, error) {
	var u = uint64(0)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v BoolVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v.Uint())
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v IntVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v Int8Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v Int16Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v Int32Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v UintVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v Uint8Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v Uint16Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v Uint32Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v FltVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v Flt32Val) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v ImagVal) MarshalBinary() ([]byte, error) {
	var n, i = uint64(real(v)), uint64(imag(v))
	var buf0 = make([]byte, 0, binary.Size(n))
	var buf1 = make([]byte, 0, binary.Size(i))
	binary.PutUvarint(buf0, n)
	binary.PutUvarint(buf1, i)
	return append(buf0, buf1...), nil
}

func (v Imag64Val) MarshalBinary() ([]byte, error) {
	var n, i = uint64(real(v)), uint64(imag(v))
	var buf0 = make([]byte, 0, binary.Size(n))
	var buf1 = make([]byte, 0, binary.Size(i))
	binary.PutUvarint(buf0, n)
	binary.PutUvarint(buf1, i)
	return append(buf0, buf1...), nil
}

func (v ByteVal) MarshalBinary() ([]byte, error) {
	var buf = make([]byte, 0, binary.Size(v))
	binary.PutUvarint(buf, uint64(v))
	return buf, nil
}

func (v BytesVal) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(v)))
	err := binary.Write(buf, binary.LittleEndian, v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (v RuneVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}

func (v StrVal) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(v)))
	err := binary.Write(buf, binary.LittleEndian, []byte(v))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (v ErrorVal) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(v.String())))
	err := binary.Write(buf, binary.LittleEndian, []byte(v.String()))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (v BigIntVal) MarshalBinary() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size((*big.Int)(&v).Bytes())))
	err := binary.Write(buf, binary.LittleEndian, (*big.Int)(&v).Bytes())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (v BigFltVal) MarshalBinary() ([]byte, error) {
	var u, _ = (*big.Float)(&v).Uint64()
	var buf = bytes.NewBuffer(make([]byte, 0, binary.Size(u)))
	err := binary.Write(buf, binary.LittleEndian, u)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (v RatioVal) MarshalBinary() ([]byte, error) {
	var d, n = uint64((*big.Rat)(&v).Denom().Uint64()), uint64((*big.Rat)(&v).Num().Uint64())
	var buf0 = make([]byte, 0, binary.Size(d))
	var buf1 = make([]byte, 0, binary.Size(n))
	binary.PutUvarint(buf0, d)
	binary.PutUvarint(buf1, n)
	return append(buf0, buf1...), nil
}

func (v TimeVal) MarshalBinary() ([]byte, error) {
	var buf = make([]byte, 0, binary.Size(v))
	(*time.Time)(&v).UnmarshalBinary(buf)
	return buf, nil
}

func (v DuraVal) MarshalBinary() ([]byte, error) {
	var u = uint64(v)
	var buf = make([]byte, 0, binary.Size(u))
	binary.PutUvarint(buf, u)
	return buf, nil
}
