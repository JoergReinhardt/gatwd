package cli

import (
	"fmt"
	"math/big"
	"strings"
	"unicode"

	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/parser"
)

// This file is the gradual-typing slice of `gatwd --check` (Workstream D1/D2 of
// docs/language_roadmap.md): optional type annotations on `def`, plus a best-effort,
// local, advisory pre-pass over the infix expression language.
//
// # The idea: gradual typing, not Hindley–Milner
//
// gatwd stays an interpreted, dynamically typed language: unannotated code runs
// exactly as before and annotations never change evaluation. The checker adds a
// *consistency* relation between the types it can see — the standard gradual-typing
// discipline: a type `Any` is consistent with everything, so anything the pass cannot
// determine (a runtime value, a pipeline step, a builtin without a signature) is
// `Any` and never produces a warning. Only a DEFINITE contradiction between two known
// types is reported. That is what keeps the pass free of false positives and why it
// is advisory (exit status unaffected).
//
// # What it catches
//
//   - arity: too many arguments to a def with a known signature or to a builtin;
//   - argument types: `def inc : Int -> Int` applied to a String literal;
//   - result types: a body whose inferred type contradicts the annotation;
//   - operator misuse: `"a" - 1`, `[1] * 2`, `true + 1` — decided by asking the
//     RUNTIME operator matrix (functions.EvalBinaryOp) on representative values, so
//     the checker can never disagree with the evaluator about what is defined;
//   - unbound names in expressions and lambda bodies;
//   - a `spawn T: seed` whose seed contradicts T's declared alias (`type IntVal = Int`).
//
// # Local inference (D2)
//
// Types flow from literals, operators, annotated parameters, let bindings and
// comprehension generators (`x <- [Int]` types x), bidirectionally only as far as a
// single statement: an application checks its arguments against the callee's
// parameter types and yields the callee's result type.

// Ty is a gradual type: Name with optional Args. Name "->" is a function type with
// Args [param, result] (curried, so `A -> B -> C` is `A -> (B -> C)`); "List" has one
// element-type arg; "Any" is the unknown type, consistent with every type.
type Ty struct {
	Name string
	Args []Ty
}

var (
	tyAny    = Ty{Name: "Any"}
	tyInt    = Ty{Name: "Int"}
	tyFloat  = Ty{Name: "Float"}
	tyString = Ty{Name: "String"}
	tyBool   = Ty{Name: "Bool"}
	tyChar   = Ty{Name: "Char"}
)

func tyList(elem Ty) Ty  { return Ty{Name: "List", Args: []Ty{elem}} }
func tyFn(p, r Ty) Ty    { return Ty{Name: "->", Args: []Ty{p, r}} }
func (t Ty) isFn() bool  { return t.Name == "->" && len(t.Args) == 2 }
func (t Ty) isAny() bool { return t.Name == "Any" || t.Name == "" }
func (t Ty) isNumeric() bool {
	switch t.Name {
	case "Int", "Float", "BigInt", "Uint", "Complex", "Ratio":
		return true
	}
	return false
}

func isCharOrString(t Ty) bool { return t.Name == "Char" || t.Name == "String" }

// ---- constraint names (hardening W3, D3) -----------------------------------------
//
// `Num`, `Integral`, `Fractional`, `Comparable` and `Appendable` are annotation-level
// CONSTRAINT names, the gradual-typing counterpart of Haskell's classes and Elm's
// `number`, `comparable`, `appendable`: `def sum2 : Num -> Num -> Num = \a b -> a + b`
// documents and checks that sum2 takes any numbers, with no runtime wrapper (nothing is
// boxed; annotations never change evaluation). A constraint names a SET of concrete types:
//
//	Num         Int, BigInt, Uint, Float, Complex, Ratio — any number
//	Integral    Int, BigInt, Uint — the operands of // and %
//	Fractional  Ratio, Float, Complex — the operands of /
//	Comparable  every type but Complex (no order on a plane) and functions; a List(T) is
//	            comparable when T is
//	Appendable  String, Bytes, List, Chunk, Set, HAMT, Tree, Graph, Pair, Maybe — what ++ joins
//
// The checker stays gradual and advisory: a value of a concrete type satisfies a constraint
// when it is in the set, a constraint-typed value is consistent with a type when SOME member
// is, and two constraints are consistent when their sets meet. Because the numeric tower
// converts at runtime (an Int promotes to a Float, `//` floors a Float), a concrete number
// is consistent with every NUMERIC constraint — so Integral and Fractional contradict
// only non-numbers and each other; they exist to state intent. Only a definite
// contradiction is reported.

var constraintNames = map[string]bool{"Num": true, "Integral": true, "Fractional": true, "Comparable": true, "Appendable": true}

// IsConstraintName reports whether name is one of the constraint names.
func IsConstraintName(name string) bool { return constraintNames[name] }

func (t Ty) isConstraint() bool { return len(t.Args) == 0 && constraintNames[t.Name] }

// numericConstraint reports whether c ranges over numbers only.
func numericConstraint(c string) bool { return c == "Num" || c == "Integral" || c == "Fractional" }

// constraintUniverse is every concrete type name a constraint can stand for; the
// consistency and operator rules are decided by enumerating it.
var constraintUniverse = []string{
	"Int", "BigInt", "Uint", "Float", "Complex", "Ratio",
	"String", "Char", "Bool", "Bytes", "Duration", "Time", "Date",
	"List", "Chunk", "Set", "HAMT", "Tree", "Graph", "Pair", "Maybe",
}

// satisfies reports whether the concrete type t is a member of constraint c.
func satisfies(c, t Ty) bool {
	if t.isAny() {
		return true
	}
	switch c.Name {
	case "Num":
		return t.isNumeric()
	case "Integral":
		return t.Name == "Int" || t.Name == "BigInt" || t.Name == "Uint"
	case "Fractional":
		return t.Name == "Ratio" || t.Name == "Float" || t.Name == "Complex"
	case "Comparable":
		if t.Name == "Complex" || t.isFn() {
			return false
		}
		if t.Name == "List" && len(t.Args) == 1 {
			return satisfies(c, t.Args[0])
		}
		return true
	case "Appendable":
		switch t.Name {
		case "String", "Bytes", "List", "Chunk", "Set", "HAMT", "Tree", "Graph", "Pair", "Maybe":
			return true
		}
	}
	return false
}

// members lists the concrete types a constraint stands for (one Ty per universe name).
func members(c Ty) []Ty {
	var out []Ty
	for _, n := range constraintUniverse {
		if t := (Ty{Name: n}); satisfies(c, t) {
			out = append(out, t)
		}
	}
	return out
}

// constraintConsistent decides consistency when at least one side is a constraint.
func constraintConsistent(want, got Ty) bool {
	switch {
	case want.isConstraint() && got.isConstraint():
		for _, m := range members(want) {
			if satisfies(got, m) {
				return true
			}
		}
		return false
	case want.isConstraint():
		return satisfies(want, got) || (numericConstraint(want.Name) && got.isNumeric())
	default:
		for _, m := range members(got) {
			if consistent(want, m) {
				return true
			}
		}
		return false
	}
}

// String renders the type in annotation syntax.
func (t Ty) String() string {
	switch {
	case t.isFn():
		l := t.Args[0].String()
		if t.Args[0].isFn() {
			l = "(" + l + ")"
		}
		return l + " -> " + t.Args[1].String()
	case len(t.Args) > 0:
		parts := make([]string, len(t.Args))
		for idx, a := range t.Args {
			parts[idx] = a.String()
		}
		return t.Name + "(" + strings.Join(parts, ", ") + ")"
	case t.Name == "":
		return "Any"
	}
	return t.Name
}

// arity counts the arrows along the result spine: Int -> Int -> Bool has arity 2.
func (t Ty) arity() int {
	n := 0
	for t.isFn() {
		n++
		t = t.Args[1]
	}
	return n
}

// consistent is the gradual-typing consistency relation: Any matches anything, all
// numeric types match each other (the numeric tower converts at runtime), and
// otherwise names and arguments must agree. It is deliberately permissive — a false
// "consistent" costs nothing, a false "inconsistent" is a false positive.
func consistent(want, got Ty) bool {
	if want.isAny() || got.isAny() {
		return true
	}
	if want.isConstraint() || got.isConstraint() {
		return constraintConsistent(want, got)
	}
	if want.isNumeric() && got.isNumeric() {
		return true
	}
	if isCharOrString(want) && isCharOrString(got) {
		return true // a one-character String converts to a Char on demand (D1)
	}
	if want.Name != got.Name {
		return false
	}
	if len(want.Args) != len(got.Args) {
		return true
	}
	for idx := range want.Args {
		if !consistent(want.Args[idx], got.Args[idx]) {
			return false
		}
	}
	return true
}

// ParseTypeAnnotation parses annotation syntax: `Int`, `List(String)`,
// `Int -> Int -> Bool`, `(Int -> Int) -> List(Int)`. `->` is right-associative.
func ParseTypeAnnotation(src string) (Ty, error) {
	p := &tyParser{src: []rune(strings.TrimSpace(src))}
	t, err := p.arrow()
	if err != nil {
		return Ty{}, err
	}
	p.skip()
	if p.pos != len(p.src) {
		return Ty{}, fmt.Errorf("unexpected %q in type annotation", string(p.src[p.pos:]))
	}
	return t, nil
}

type tyParser struct {
	src []rune
	pos int
}

func (p *tyParser) skip() {
	for p.pos < len(p.src) && unicode.IsSpace(p.src[p.pos]) {
		p.pos++
	}
}

func (p *tyParser) arrow() (Ty, error) {
	left, err := p.app()
	if err != nil {
		return Ty{}, err
	}
	p.skip()
	if p.pos+1 < len(p.src) && p.src[p.pos] == '-' && p.src[p.pos+1] == '>' {
		p.pos += 2
		right, err := p.arrow()
		if err != nil {
			return Ty{}, err
		}
		return tyFn(left, right), nil
	}
	return left, nil
}

func (p *tyParser) app() (Ty, error) {
	p.skip()
	if p.pos >= len(p.src) {
		return Ty{}, fmt.Errorf("expected a type")
	}
	if p.src[p.pos] == '(' {
		p.pos++
		t, err := p.arrow()
		if err != nil {
			return Ty{}, err
		}
		p.skip()
		if p.pos >= len(p.src) || p.src[p.pos] != ')' {
			return Ty{}, fmt.Errorf("missing ')' in type annotation")
		}
		p.pos++
		return t, nil
	}
	start := p.pos
	for p.pos < len(p.src) && (unicode.IsLetter(p.src[p.pos]) || unicode.IsDigit(p.src[p.pos]) || p.src[p.pos] == '_') {
		p.pos++
	}
	if start == p.pos {
		return Ty{}, fmt.Errorf("expected a type name at %q", string(p.src[start:]))
	}
	t := Ty{Name: string(p.src[start:p.pos])}
	if constraintNames[t.Name] && p.pos < len(p.src) && p.src[p.pos] == '(' {
		return Ty{}, fmt.Errorf("the constraint %s takes no type arguments", t.Name)
	}
	if p.pos < len(p.src) && p.src[p.pos] == '(' {
		p.pos++
		for {
			arg, err := p.arrow()
			if err != nil {
				return Ty{}, err
			}
			t.Args = append(t.Args, arg)
			p.skip()
			if p.pos < len(p.src) && p.src[p.pos] == ',' {
				p.pos++
				continue
			}
			break
		}
		if p.pos >= len(p.src) || p.src[p.pos] != ')' {
			return Ty{}, fmt.Errorf("missing ')' in type annotation")
		}
		p.pos++
	}
	return t, nil
}

// splitDefHead splits the left side of `def <head> = body` into the function name
// and its optional `: type` annotation, so `def inc : Int -> Int = …` and the plain
// `def inc = …` share one reader. The annotation never affects evaluation.
func splitDefHead(head string) (name, sig string) {
	if idx := strings.Index(head, ":"); idx >= 0 {
		return strings.TrimSpace(head[:idx]), strings.TrimSpace(head[idx+1:])
	}
	return strings.TrimSpace(head), ""
}

// ---- value → type, and the operator oracle ------------------------------------

func tyOfValue(v any) Ty {
	switch x := v.(type) {
	case bool:
		return tyBool
	case string:
		return tyString
	case functions.Char:
		return tyChar
	case functions.Bytes, []byte:
		return Ty{Name: "Bytes"}
	case float32, float64:
		return tyFloat
	case complex64, complex128:
		return Ty{Name: "Complex"}
	case *big.Int:
		return Ty{Name: "BigInt"}
	case *functions.Ratio:
		return Ty{Name: "Ratio"}
	case functions.Duration:
		return Ty{Name: "Duration"}
	case functions.Time:
		return Ty{Name: "Time"}
	case functions.Date:
		return Ty{Name: "Date"}
	case *collections.List[any]:
		return tyList(tyAny)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return tyInt
	default:
		_ = x
		return tyAny
	}
}

// sampleOf returns a representative runtime value of t, used to ask the evaluator's
// own operator matrix whether an operation is defined.
func sampleOf(t Ty) (any, bool) {
	switch t.Name {
	case "Int", "Uint":
		return 1, true
	case "Float":
		return 1.5, true
	case "String":
		return "s", true
	case "Bool":
		return true, true
	case "Char":
		return functions.Char('a'), true
	case "BigInt":
		return big.NewInt(1), true
	case "Ratio":
		r, _ := functions.NewRatio(big.NewInt(1), big.NewInt(2))
		return r, true
	case "Complex":
		return complex(1, 1), true
	case "Duration":
		return functions.Duration(24 * 3600e9), true // a whole day, so `Date + Duration` is defined
	case "Time":
		t, _ := functions.ParseTime("2026-10-13T09:30:00Z")
		return t, true
	case "Date":
		d, _ := functions.NewDate(2026, 10, 13)
		return d, true
	case "List":
		return collections.Cons[any](1, collections.Nil[any]()), true
	}
	return nil, false
}

// binaryType types `l op r`. When both operands are concrete it evaluates the real
// operator on samples: an error means the operator is genuinely undefined for the
// pair (reported), success yields the result type. Anything involving Any is Any.
func binaryType(op functions.OpKind, l, r Ty) (Ty, string) {
	if l.isConstraint() || r.isConstraint() {
		return constraintBinaryType(op, l, r)
	}
	if l.isAny() || r.isAny() {
		switch op {
		case functions.OpEq, functions.OpNeq, functions.OpLt, functions.OpLte, functions.OpGt, functions.OpGte, functions.OpAnd, functions.OpOr:
			return tyBool, ""
		}
		return tyAny, ""
	}
	lv, lok := sampleOf(l)
	rv, rok := sampleOf(r)
	if !lok || !rok {
		return tyAny, ""
	}
	var res any
	var err error
	if op == functions.OpConcat {
		res, err = parser.ConcatValues(lv, rv)
	} else {
		res, err = functions.EvalBinaryOp(op, lv, rv)
	}
	if err != nil {
		return tyAny, fmt.Sprintf("operator '%s' is not defined for %s and %s", op, l, r)
	}
	t := tyOfValue(res)
	if t.Name == "List" {
		t = tyList(joinTy(elemTy(l), elemTy(r)))
	}
	// A fraction that happens to be whole demotes to an Int, so the sample answered Int
	// for `Ratio + Ratio`; the type of the operation is the Ratio the operands generally give.
	if t.Name == "Int" && (l.Name == "Ratio" || r.Name == "Ratio") {
		switch op {
		case functions.OpAdd, functions.OpSub, functions.OpMul, functions.OpDiv, functions.OpPow:
			t = Ty{Name: "Ratio"}
		}
	}
	// Duration / Duration is the exact ratio (an Int only when the sample happened to divide).
	if t.Name == "Int" && l.Name == "Duration" && r.Name == "Duration" && op == functions.OpDiv {
		t = Ty{Name: "Ratio"}
	}
	return t, ""
}

// constraintBinaryType types `l op r` when an operand is a constraint: the operator is
// asked about every pair of member types. If it is defined for NO pair the operation is a
// definite error; otherwise the result is the common type of the pairs that work (all Int
// → Int, all numbers → Num, anything else → Any, comparisons → Bool).
func constraintBinaryType(op functions.OpKind, l, r Ty) (Ty, string) {
	if l.isAny() || r.isAny() {
		return binaryType(op, tyAny, tyAny)
	}
	ls, rs := []Ty{l}, []Ty{r}
	if l.isConstraint() {
		ls = members(l)
	}
	if r.isConstraint() {
		rs = members(r)
	}
	var results []Ty
	unknown := false // a member the oracle has no sample for might support the operator
	for _, a := range ls {
		for _, b := range rs {
			_, aok := sampleOf(a)
			_, bok := sampleOf(b)
			if !aok || !bok {
				unknown = true
				continue
			}
			if t, msg := binaryType(op, a, b); msg == "" {
				results = append(results, t)
			}
		}
	}
	switch {
	case len(results) > 0 && !unknown:
		return joinResults(results), ""
	case len(results) > 0:
		return tyAny, ""
	case unknown:
		return tyAny, ""
	}
	return tyAny, fmt.Sprintf("operator '%s' is not defined for %s and %s", op, l, r)
}

// joinResults is the common type of operator results: equal types stay, a mix of
// numbers is Num, anything else is Any.
func joinResults(ts []Ty) Ty {
	out := ts[0]
	for _, t := range ts[1:] {
		switch {
		case t.String() == out.String():
		case t.isNumeric() && out.isNumeric():
			out = Ty{Name: "Num"}
		case t.Name == "Num" && out.isNumeric(), out.Name == "Num" && t.isNumeric():
			out = Ty{Name: "Num"}
		default:
			return tyAny
		}
	}
	return out
}

// inexactIntBound is 2^53: beyond it an Int is not exactly representable as a Float.
var inexactIntBound = new(big.Int).Lsh(big.NewInt(1), 53)

// warnInexactIntToFloat is the advisory of the promotion lattice (plan §6 item 2): an Int
// LITERAL larger than 2^53 combined with a Float converts to the nearest double and loses
// digits silently. `to_float` states the conversion; keeping the operation on integers
// avoids it. Only a literal is judged (the checker does not track values), so there are no
// false positives.
func (c *typeChecker) warnInexactIntToFloat(x parser.BinaryExpr, l, r Ty) {
	switch x.Op {
	case functions.OpAdd, functions.OpSub, functions.OpMul, functions.OpDiv, functions.OpPow:
	default:
		return
	}
	check := func(e parser.Expr, other Ty) {
		lit, ok := e.(parser.LiteralExpr)
		if !ok || other.Name != "Float" {
			return
		}
		b, isInt := functions.ToBigInt(lit.Value)
		if !isInt || new(big.Int).Abs(b).Cmp(inexactIntBound) <= 0 {
			return
		}
		c.warn(fmt.Sprintf("integer %s is larger than 2^53 and converts to Float inexactly in '%s'; use to_float to say so, or keep the operation on integers", b, x.Op))
	}
	check(x.Left, r)
	check(x.Right, l)
}

// unaryType types a prefix operator by asking the runtime (parser.ApplyUnary) about
// representative values, like binaryType: an error is a definite contradiction.
func unaryType(op string, t Ty) (Ty, string) {
	if t.isAny() {
		if op == "!" {
			return tyBool, "" // `!x` is a Bool whatever x turns out to be (or an error)
		}
		return tyAny, ""
	}
	cands := []Ty{t}
	if t.isConstraint() {
		cands = members(t)
	}
	var results []Ty
	for _, cand := range cands {
		v, ok := sampleOf(cand)
		if !ok {
			return tyAny, ""
		}
		if res, err := parser.ApplyUnary(op, v); err == nil {
			results = append(results, tyOfValue(res))
		}
	}
	if len(results) == 0 {
		return tyAny, fmt.Sprintf("operator '%s' is not defined for %s", op, t)
	}
	return joinResults(results), ""
}

func elemTy(t Ty) Ty {
	if t.Name == "List" && len(t.Args) == 1 {
		return t.Args[0]
	}
	return tyAny
}

// joinTy is the least upper bound used for list elements: equal types stay, anything
// else widens to Any.
func joinTy(a, b Ty) Ty {
	if a.String() == b.String() {
		return a
	}
	if a.isNumeric() && b.isNumeric() {
		// the join in the tower Int ⊂ Ratio ⊂ Float ⊂ Complex
		for _, name := range []string{"Complex", "Float", "Ratio"} {
			if a.Name == name || b.Name == name {
				return Ty{Name: name}
			}
		}
		return tyInt
	}
	return tyAny
}

// ---- the checker -------------------------------------------------------------------

// Diagnostic is one finding of CheckScript.
type Diagnostic struct {
	Line     int    // 1-based line where the statement began
	Col      int    // 1-based column (1 when unknown)
	Severity string // "error" (structural) or "warning" (advisory)
	Msg      string
}

type tyScope map[string]Ty

func (s tyScope) with(name string, t Ty) tyScope {
	out := make(tyScope, len(s)+1)
	for k, v := range s {
		out[k] = v
	}
	out[name] = t
	return out
}

type typeChecker struct {
	defs     tyScope           // def/const/let-introduced top-level names
	aliases  map[string]Ty     // `type IntVal = Int`
	ctors    func(string) bool // declared sum-type constructors
	warnLine func(msg string)
	seen     map[string]bool
	// lenient > 0 while inside an application's head/arguments: gatwd lambda bodies
	// may be pipeline-style (`echo hello`, `focus Name`) where bare words are strings
	// or step names, so an unresolved name there is not reported. A lambda body or an
	// operator operand is a plain expression context and resets it.
	lenient int
}

func (c *typeChecker) warn(msg string) {
	if c.seen[msg] {
		return
	}
	c.seen[msg] = true
	c.warnLine(msg)
}

func (c *typeChecker) resolve(t Ty) Ty {
	if a, ok := c.aliases[t.Name]; ok && len(t.Args) == 0 {
		return a
	}
	for idx := range t.Args {
		t.Args[idx] = c.resolve(t.Args[idx])
	}
	return t
}

func builtinArity(name string) (int, bool) {
	v, ok := dslBuiltinLookup(name)
	if !ok {
		return 0, false
	}
	if cl, isClosure := v.(*functions.Closure); isClosure {
		return cl.Arity(), true
	}
	return 0, false
}

func (c *typeChecker) known(name string, scope tyScope) (Ty, bool) {
	if t, ok := scope[name]; ok {
		return t, true
	}
	if t, ok := c.defs[name]; ok {
		return t, true
	}
	if _, ok := dslBuiltinLookup(name); ok {
		return tyAny, true
	}
	if name == "true" || name == "false" || name == "variant" {
		return tyBool, true
	}
	if name == "now" {
		return Ty{Name: "Time"}, true // the clock step's reading, usable as a pipeline head (D15)
	}
	if c.ctors != nil && c.ctors(name) {
		return tyString, true
	}
	return Ty{}, false
}

// infer types an expression within scope, warning on definite contradictions.
func (c *typeChecker) infer(e parser.Expr, scope tyScope) Ty {
	switch x := e.(type) {
	case parser.LiteralExpr:
		return tyOfValue(x.Value)
	case parser.VarExpr:
		t, ok := c.known(x.Name, scope)
		if !ok {
			if c.lenient == 0 {
				c.warn(fmt.Sprintf("unbound name '%s'", x.Name))
			}
			return tyAny
		}
		if x.Name == "true" || x.Name == "false" {
			return tyBool
		}
		return t
	case parser.BinaryExpr:
		saved := c.lenient
		c.lenient = 0
		l := c.infer(x.Left, scope)
		r := tyAny
		// `Computed_%s`: a printf verb in step text is not a modulo operand.
		if v, isVar := x.Right.(parser.VarExpr); !(isVar && x.Op == functions.OpMod && len(v.Name) == 1 && strings.Contains("sdvqfxt", v.Name)) {
			r = c.infer(x.Right, scope)
		}
		c.lenient = saved
		t, msg := binaryType(x.Op, l, r)
		if msg != "" {
			c.warn(msg)
		}
		c.warnInexactIntToFloat(x, l, r)
		return t
	case parser.UnaryExpr:
		saved := c.lenient
		c.lenient = 0
		t := c.infer(x.Operand, scope)
		c.lenient = saved
		res, msg := unaryType(x.Op, t)
		if msg != "" {
			c.warn(msg)
		}
		return res
	case parser.ListLiteralExpr:
		elem := Ty{}
		for idx, el := range x.Elements {
			t := c.infer(el, scope)
			if idx == 0 {
				elem = t
			} else {
				elem = joinTy(elem, t)
			}
		}
		if len(x.Elements) == 0 {
			elem = tyAny
		}
		return tyList(elem)
	case parser.ListComprehensionExpr:
		inner := scope
		for _, g := range x.Generators {
			src := c.infer(g.Source, inner)
			elem := elemTy(src)
			for _, n := range g.PatternNames {
				if len(g.PatternNames) == 1 {
					inner = inner.with(n, elem)
				} else {
					inner = inner.with(n, tyAny)
				}
			}
		}
		for _, f := range x.Filters {
			c.infer(f, inner)
		}
		return tyList(c.infer(x.YieldExpr, inner))
	case parser.LambdaExpr:
		inner := scope
		for _, p := range x.Params {
			inner = inner.with(p, tyAny)
		}
		saved := c.lenient
		c.lenient = 0
		res := c.infer(x.Body, inner)
		c.lenient = saved
		t := res
		for idx := len(x.Params) - 1; idx >= 0; idx-- {
			t = tyFn(tyAny, t)
		}
		return t
	case parser.ApplyExpr:
		return c.inferApply(x, scope)
	}
	return tyAny
}

func (c *typeChecker) inferApply(x parser.ApplyExpr, scope tyScope) Ty {
	c.lenient++
	head := c.infer(x.Head, scope)
	argTys := make([]Ty, len(x.Args))
	for idx, a := range x.Args {
		argTys[idx] = c.infer(a, scope)
	}
	c.lenient--
	name := ""
	if v, ok := x.Head.(parser.VarExpr); ok {
		name = v.Name
	}
	// A builtin's arity is known from its curried registration; no builtin returns a
	// function once saturated, so extra arguments are a definite error.
	if _, isDef := c.defs[name]; name != "" && !isDef {
		if _, isLocal := scope[name]; !isLocal {
			if ar, ok := builtinArity(name); ok && len(x.Args) > ar {
				c.warn(fmt.Sprintf("'%s' takes %d argument(s) but is applied to %d", name, ar, len(x.Args)))
			}
			return tyAny
		}
	}
	return c.applyType(name, head, argTys)
}

// applyType checks argTys against the parameter spine of fn and returns the type left
// after applying them.
func (c *typeChecker) applyType(name string, fn Ty, args []Ty) Ty {
	label := name
	if label == "" {
		label = "function"
	}
	cur := fn
	for idx, a := range args {
		if cur.isAny() {
			return tyAny
		}
		if !cur.isFn() {
			c.warn(fmt.Sprintf("'%s' takes %d argument(s) but is applied to %d", label, fn.arity(), len(args)))
			return tyAny
		}
		if !consistent(cur.Args[0], a) {
			c.warn(fmt.Sprintf("argument %d of '%s' has type %s, expected %s", idx+1, label, a, cur.Args[0]))
		}
		cur = cur.Args[1]
	}
	return cur
}

// ---- statements ------------------------------------------------------------------

// lambdaGroups returns the parenthesised `(\…)` groups in step text, so lambdas
// nested in pipeline steps (`map (\x -> …)`, `fold (\g r -> …) seed`) are checked
// with the same expression checker as top-level expressions.
func lambdaGroups(s string) []string {
	var out []string
	for idx := 0; idx < len(s); idx++ {
		if s[idx] == '"' {
			for idx++; idx < len(s) && s[idx] != '"'; idx++ {
				if s[idx] == '\\' {
					idx++ // an escape: skip the escaped byte
				}
			}
			continue
		}
		if s[idx] != '(' || idx+1 >= len(s) || s[idx+1] != '\\' {
			continue
		}
		depth, end := 0, -1
		for j := idx; j < len(s); j++ {
			if s[j] == '(' {
				depth++
			} else if s[j] == ')' {
				depth--
				if depth == 0 {
					end = j
					break
				}
			}
		}
		if end < 0 {
			break
		}
		out = append(out, s[idx+1:end])
		idx = end
	}
	return out
}

func (c *typeChecker) checkExprText(text string, scope tyScope) (Ty, bool) {
	expr, err := parser.ParseInfixExpr(strings.TrimSpace(text))
	if err != nil {
		return tyAny, false
	}
	return c.infer(expr, scope), true
}

// seedType infers the type of a `spawn T: seed` seed: a number, a quoted string or an
// arithmetic expression has a definite type; a bare word is a string.
func (c *typeChecker) checkSpawn(first string, scope tyScope) Ty {
	rest := strings.TrimSpace(strings.TrimPrefix(first, "spawn "))
	colon := strings.Index(rest, ":")
	if colon < 0 {
		return tyAny
	}
	typeName := strings.TrimSpace(rest[:colon])
	seed := strings.TrimSpace(rest[colon+1:])
	want, ok := c.aliases[typeName]
	if !ok {
		return tyAny
	}
	// Only scalar primitives have a seed whose type can be judged from its text; any
	// other alias (a nominal type, a collection, a plugin type) is built by a factory
	// that accepts arbitrary words, so its seed is opaque.
	if !want.isNumeric() && want.Name != "Bool" {
		return want
	}
	var got Ty
	if expr, err := parser.ParseInfixExpr(seed); err == nil {
		if v, isVar := expr.(parser.VarExpr); isVar {
			if _, bound := c.known(v.Name, scope); bound {
				return want
			}
			got = tyString
		} else {
			got = c.infer(expr, scope)
		}
	} else {
		got = tyString
	}
	if !consistent(want, got) {
		c.warn(fmt.Sprintf("spawn %s: seed '%s' has type %s, but %s is declared as %s", typeName, seed, got, typeName, want))
	}
	return want
}

// checkStatement checks one logical statement, extending c.defs for declarations.
func (c *typeChecker) checkStatement(line string) {
	switch {
	case strings.HasPrefix(line, "def "):
		c.checkDef(line[4:])
		return
	case strings.HasPrefix(line, "const "):
		parts := strings.SplitN(line[6:], "=", 2)
		if len(parts) == 2 {
			t, ok := c.checkExprText(parts[1], nil)
			if !ok {
				t = tyAny
			}
			c.defs[strings.TrimSpace(parts[0])] = t
		}
		return
	case strings.HasPrefix(line, "type "), strings.HasPrefix(line, "import "):
		return
	}

	if strings.HasPrefix(line, "let ") && !strings.Contains(line, " in ") {
		if def := strings.SplitN(line[4:], "=", 2); len(def) == 2 {
			t := tyAny
			rhs := strings.TrimSpace(def[1])
			if !strings.Contains(rhs, "|>") && !strings.HasPrefix(rhs, "spawn ") {
				if it, ok := c.checkExprText(rhs, nil); ok {
					t = it
				}
			}
			c.defs[strings.TrimSpace(def[0])] = t
		}
		return
	}

	scope := tyScope{}
	// `let x = rhs in …` chains bind x for the rest of the statement.
	for strings.HasPrefix(line, "let ") {
		parts := strings.SplitN(line[4:], " in ", 2)
		if len(parts) != 2 {
			return
		}
		def := strings.SplitN(parts[0], "=", 2)
		if len(def) != 2 {
			return
		}
		name := strings.TrimSpace(def[0])
		rhs := strings.TrimSpace(def[1])
		t := tyAny
		if !strings.Contains(rhs, "|>") && !strings.HasPrefix(rhs, "spawn ") {
			if it, ok := c.checkExprText(rhs, scope); ok {
				t = it
			}
		}
		scope = scope.with(name, t)
		line = strings.TrimSpace(parts[1])
	}

	if !strings.Contains(line, "|>") {
		if !strings.HasPrefix(line, "spawn ") {
			c.checkExprText(line, scope)
		} else {
			c.checkSpawn(line, scope)
		}
		return
	}
	steps, _ := splitPipelineAware(line)
	incoming := tyAny
	for idx, step := range steps {
		step = strings.TrimSpace(step)
		if idx == 0 {
			if strings.HasPrefix(step, "spawn ") {
				incoming = c.checkSpawn(step, scope)
			} else if t, ok := c.checkExprText(step, scope); ok {
				incoming = t
			}
			continue
		}
		for _, lam := range lambdaGroups(step) {
			c.checkExprText(lam, scope)
		}
		incoming = c.checkStep(step, incoming, scope)
	}
}

// checkStep checks a pipeline step whose head is a typed user function: arguments
// first, the incoming value LAST (data-last application). Any other step is opaque.
func (c *typeChecker) checkStep(step string, incoming Ty, scope tyScope) Ty {
	fields := strings.Fields(step)
	if len(fields) == 0 {
		return tyAny
	}
	fn, ok := c.defs[fields[0]]
	if !ok {
		fn, ok = scope[fields[0]]
	}
	if !ok || !fn.isFn() {
		return tyAny
	}
	var args []Ty
	if len(fields) > 1 {
		expr, err := parser.ParseInfixExpr(step)
		app, isApp := expr.(parser.ApplyExpr)
		if err != nil || !isApp {
			return tyAny
		}
		for _, a := range app.Args {
			args = append(args, c.infer(a, scope))
		}
	}
	if fn.arity() == 1 && len(args) == 0 {
		// A bare one-argument function maps over collections: element-wise, so the
		// incoming type is not checked against the parameter directly.
		return fn.Args[1]
	}
	return c.applyType(fields[0], fn, append(args, incoming))
}

func (c *typeChecker) checkDef(rest string) {
	parts := strings.SplitN(rest, "=", 2)
	if len(parts) != 2 {
		return
	}
	name, sigSrc := splitDefHead(parts[0])
	body := strings.TrimSpace(parts[1])

	declared := tyAny
	if sigSrc != "" {
		sig, err := ParseTypeAnnotation(sigSrc)
		if err != nil {
			c.warn(fmt.Sprintf("def %s: %v", name, err))
		} else {
			declared = c.resolve(sig)
		}
	}
	c.defs[name] = declared // visible to its own body (recursion)

	expr, err := parser.ParseInfixExpr(body)
	if err != nil {
		return
	}
	lam, isLam := expr.(parser.LambdaExpr)
	if !isLam {
		return // a non-lambda body is a pipeline macro: its words are steps, not expressions
	}
	scope := tyScope{}
	rem := declared
	for _, p := range lam.Params {
		pt := tyAny
		if rem.isFn() {
			pt, rem = rem.Args[0], rem.Args[1]
		} else if !rem.isAny() {
			c.warn(fmt.Sprintf("def %s: lambda has %d parameter(s) but the annotation %s allows fewer", name, len(lam.Params), declared))
			rem = tyAny
		}
		scope = scope.with(p, pt)
	}
	got := c.infer(lam.Body, scope)
	if !rem.isAny() && !consistent(rem, got) {
		c.warn(fmt.Sprintf("def %s: body has type %s but the annotation says %s", name, got, rem))
	}
	if declared.isAny() {
		t := got
		for idx := len(lam.Params) - 1; idx >= 0; idx-- {
			t = tyFn(tyAny, t)
		}
		c.defs[name] = t
	}
}

// CheckScript statically checks a gatwd script and returns its diagnostics: structural
// errors (legacy `:=`, unbalanced delimiters, declarations missing '=', a dangling
// `|>`) and advisory warnings (non-exhaustive `match`, the gradual-typing findings of
// this file). It never evaluates the script.
func CheckScript(content string) []Diagnostic {
	stmts := logicalLines(content)

	checkParser := parser.NewParser()
	aliases := map[string]Ty{}
	for _, stmt := range stmts {
		if !strings.HasPrefix(stmt.text, "type ") {
			continue
		}
		for range checkParser.Parse(stmt.text) {
		}
		parts := strings.SplitN(stmt.text[5:], "=", 2)
		if len(parts) == 2 && !strings.Contains(parts[1], "|") {
			if t, err := ParseTypeAnnotation(parts[1]); err == nil {
				aliases[strings.TrimSpace(parts[0])] = t
			}
		}
	}

	var diags []Diagnostic
	tc := &typeChecker{defs: tyScope{}, aliases: aliases, ctors: checkParser.IsConstructor, seen: map[string]bool{}}
	for _, stmt := range stmts {
		line := stmt.text
		report := func(col int, msg string) {
			diags = append(diags, Diagnostic{Line: stmt.num, Col: col, Severity: "error", Msg: msg})
		}
		warn := func(msg string) {
			diags = append(diags, Diagnostic{Line: stmt.num, Col: 1, Severity: "warning", Msg: msg})
		}
		checkExhaustiveMatch(checkParser, line, warn)
		if strings.Contains(line, ":=") && !strings.Contains(line, "ebnf_parse") {
			report(strings.Index(line, ":=")+1, "invalid assignment ':=' (use '=')")
		}
		structural := false
		if err := parser.CheckComprehensions(line); err != nil {
			// Not advisory: evaluating the form is always an error, so it is reported as one.
			report(1, err.Error())
		}
		if col, ok := unbalancedDelimiter(line); !ok {
			report(col+1, "unbalanced delimiter")
			structural = true
		}
		for _, kw := range []string{"def ", "const ", "type "} {
			if strings.HasPrefix(line, kw) && !strings.Contains(line, "=") {
				report(1, "'"+strings.TrimSpace(kw)+"' declaration is missing '='")
				structural = true
			}
		}
		if strings.HasSuffix(line, "|>") {
			report(len(line)-1, "pipeline ends with '|>' but has no following stage")
			structural = true
		}
		if !structural {
			tc.warnLine = func(msg string) { warn(msg) }
			tc.seen = map[string]bool{}
			tc.checkStatement(line)
		}
	}
	return diags
}
