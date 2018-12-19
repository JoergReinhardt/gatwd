package types

// functions to implement the value interface on arbitrary types during runtime
func FlagFn(v Value) Flag       { return v.Flag() }
func TypeFn(v Value) Type       { return v.Type() }
func CopyFn(v Value) Value      { return v.Copy() }
func RefFn(v Value) interface{} { return v.Ref() }
func DeRefFn(v Value) Value     { return v.DeRef() }
