package functions

import (
	"fmt"
	"testing"

	d "github.com/joergreinhardt/gatwd/data"
)

func TestCase(t *testing.T) {
	var p1, _ = NewPredictNarg(func(args ...Callable) bool {
		for _, arg := range args {
			if !arg.(Native).SubType().Match(d.String | d.Integers | d.Float) {
				return false
			}
		}
		return true
	}),
		NewPredictNarg(func(args ...Callable) bool {
			for _, arg := range args {
				if !arg.(Native).SubType().Match(d.String | d.Integers | d.Float) {
					return false
				}
			}
			return true
		})

	var c = NewCase(p1,
		NewNary(func(args ...Callable) Callable {
			var nats = []d.Native{}
			var expr Callable
			if len(args) > 0 {
				expr = args[0]
				if len(args) > 1 {
					args = args[1:]
				}
			}
			for _, arg := range args {
				nats = append(nats, arg.Eval())
			}
			return expr.Call(NewNative(nats...))
		}, 2),
		NewVariadic(UnaryLambda(func(arg Callable) Callable {
			return NewNone()
		})),
	)

	fmt.Printf("case type name: %s\n", c.TypeName())

	fmt.Printf("case pred nary two expressions expected false: %s\n",
		c.Call(New("string"), New([]byte("bytes"))))

	if c.Call(New("string"), New([]byte("bytes"))).Call().(Native)().(d.BoolVal) {
		t.Fail()
	}

	fmt.Printf("case pred all two expressions expected true: %s\n",
		c.Call(New("string"), New(10)))

	if !c.Call(New("string"), New(10)).Call().(Native)().(d.BoolVal) {
		t.Fail()
	}
}

func TestSwitch(t *testing.T) {
	var swi = NewSwitch(
		// matches return values native types string,integer, and float
		NewCase(NewPredictAll(func(arg Callable) bool {
			return arg.(SubTyped).SubType().Match(d.String | d.Integers | d.Float)
		})))

	fmt.Printf("switch int & float argument: %s\n", swi.Call(New(23), New(42, 23)))

	fmt.Printf("successfull call to Switch passing int: %s\n", swi.Call(New(42)))

	if val := swi.Call(New(42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call to Switch passing float: %s\n", swi.Call(New(23.42)))
	if val := swi.Call(New(23.42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call to Switch passing string: %s\n", swi.Call(New("string")))
	if val := swi.Call(New(23.42)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call to Switch passing multiple integers: %s\n",
		swi.Call(New(23), New(42), New(65)))
	if val := swi.Call(New(23), New(42), New(65)); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("successfull call to Switch passing mixed args: %s\n",
		swi.Call(New(23), New(42.23), New("string")))
	if val := swi.Call(New(23), New(42.23), New("string")); val.TypeFnc().Match(None) {
		t.Fail()
	}

	fmt.Printf("unsuccessfull call to Switch passing boolean: %s\n", swi.Call(New(true)))
	if val := swi.Call(New(true)); !val.TypeFnc().Match(None) {
		t.Fail()
	}
}

func TestMaybe(t *testing.T) {

	var maybe = NewMaybe(NewCase(NewPredictArg(func(arg Callable) bool {
		return arg.(Native).SubType().Match(d.String)
	}).Nargs(), NewUnary(func(arg Callable) Callable { return arg })))

	fmt.Printf("maybe: %s\n", maybe)

	var str = maybe(New("string"))
	fmt.Printf("str: %s str type name: %s\n", str, str.TypeName())
	if str.String() != "string" {
		t.Fail()
	}

	var none = maybe(New(1))
	fmt.Printf("none: %s none type name: %s\n", none, none.TypeName())

	if none.TypeFnc() != None {
		t.Fail()
	}
}

func TestEither(t *testing.T) {
	var either = NewEither(NewCase(
		NewPredictArg(
			func(arg Callable) bool {
				return arg.(Native).SubType().Match(d.String)
			}).Nargs()),
		nil,
		NewUnary(func(arg Callable) Callable {
			return NewNative(d.ErrorVal{fmt.Errorf("error: " + arg.String())})
		}),
	)
	fmt.Printf("either: %s either type name: %s\n", either, either.TypeName())

	var str = either(New("string"))
	fmt.Printf("str: %s str type name: %s fnc type: %s nat type: %s\n",
		str, str.TypeName(), str.TypeFnc(), str.TypeNat().TypeName())
	if str.String() != "string" {
		t.Fail()
	}

	var err = either(New(1))
	fmt.Printf("err: %s err type name: %s type nat: %s\n", err, err.TypeName(), err.TypeNat())
	if err.Eval().(d.ErrorVal).E.Error() != "error: 1" {
		t.Fail()
	}
}
