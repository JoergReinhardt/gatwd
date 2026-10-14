# gatwd Formal Grammar (EBNF)

This is the formal Extended Backus-Naur Form (EBNF) grammar for the `gatwd` Domain Specific Language.

```ebnf
Program        ::= { Statement }
Statement      ::= Import | TypeDef | ConstDef | FunctionDef | LetBinding | Pipeline | Comment
Comment        ::= "--" { ANY_CHAR } "\n"

(* Modularity & Reusability *)
Import         ::= "import " StringLiteral
FunctionDef    ::= "def " Identifier [ ":" TypeAnnotation ] "=" ( Pipeline | LambdaExpr )   (* the annotation is advisory: read by `gatwd --check`, ignored by evaluation *)
TypeAnnotation ::= TypeAtom { "->" TypeAtom }                                  (* right-associative; "(" TypeAnnotation ")" groups *)
TypeAtom       ::= Identifier [ "(" TypeAnnotation { "," TypeAnnotation } ")" ]
ConstDef       ::= "const " Identifier "=" Pipeline                          (* a script-wide constant; ARG0 / ARGS are predefined *)
LetBinding     ::= "let " Identifier "=" Pipeline " in " Pipeline

(* Type Definitions (Algebraic Data Types) *)
TypeDef        ::= "type " Identifier "=" (SumType | RecordType | TupleType | CollectionType | NativeType)
NativeType     ::= Identifier (* Maps to Go Primitives registered via Interpreter.RegisterTypeFactory *)
SumType        ::= Identifier { "|" Identifier }
RecordType     ::= "{" FieldDef { "," FieldDef } "}"
TupleType      ::= "(" Identifier { "," Identifier } ")"
CollectionType ::= ( "Set(" | "List(" | "Graph(" | "Tree(" | "BST(" | "BTree(" | "RBTree(" ) Identifier ")"
FieldDef       ::= Identifier ":" Identifier

(* Pipelines (Graph Traversal & Evaluation) *)
Pipeline       ::= SpawnNode { ( "|>" | "||>" ) PipelineStep }
SpawnNode      ::= "spawn " Identifier ":" Identifier

(* Core Mechanics & Transducers *)
PipelineStep   ::= OSCommand 
                 | "take " Integer
                 | "drop " Integer
                 | "fold " FoldOp
                 | "distinct"
                 | "memoize " OSCommand
                 | "map " Lambda
                 | Lambda                         (* bare lambda applies to the incoming value *)
                 | FuncName                        (* a def, incl. lambda-bodied, applied by name *)
                 | FuncName Atom { Atom }          (* a closure applied to arguments; the incoming value is the LAST argument *)
                 | "fold " Atom [ Atom ]           (* fold a 2-ary closure (name or partial application) with optional seed *)
                 | "filter " Atom | "reject " Atom (* a predicate closure instead of an inline lambda *)
                 | LetBinding                      (* let x = Pipeline in Pipeline, as a step *)

FoldOp         ::= "count" | "sum" | "product" | "min" | "max" | "concat" | "any" | "all"
                 | Lambda Atom                    (* custom left fold: fold (\acc x -> expr) seed *)
Lambda         ::= [ "(" ] "\\" Identifier { Identifier } "->" Pipeline [ ")" ]   (* several parameters make a curried closure *)

(* Error Handling & Compliance *)
                 | "expect " StringLiteral " " StringLiteral
                 | "filter " StringLiteral

(* Arrow Routing (Kleisli & Applicative) *)
                 | "(" Pipeline { " <+> " Pipeline } ")"  (* ArrowPlus: run the incoming node through every branch, concatenate outputs in order *)
                 | "(" OSCommand " &&& " OSCommand ")"
                 | "[" OSCommand " *** " OSCommand "]"
                 | "zip " OSCommand
                 | "format_tuple " StringLiteral

(* Pattern Matching (Catamorphisms) *)
                 | "match {" MatchRule { "|" MatchRule } "}"
MatchRule      ::= MatchPattern "->" Pipeline      (* the pattern ends at the first "->" outside a String/Char literal *)
MatchPattern   ::= Constructor                      (* registered sum-type branch: discriminates *)
                 | "_"                              (* wildcard catch-all *)
                 | Identifier                        (* fresh variable: binds any value *)
                 | "[]"                             (* empty list *)
                 | "[" ElemPattern ", ..." Identifier "]"   (* head / tail: ONE head element pattern, a tail name or _ *)
                 | "[" ElemPattern { "," ElemPattern } "]" (* fixed-length list/tuple; also "(" ... ")" *)
                 | Literal                          (* equality, never substring: see below *)
Arm            ::= MatchPattern [ "if" Expr ] "->" Expr   (* guard: false falls through to the next arm *)
LambdaParam    ::= Identifier | MatchPattern   (* a non-identifier parameter makes the lambda a hoisted match *)
ClauseDef      ::= "def" Identifier MatchPattern { MatchPattern } [ "if" Expr ] "=" Expr   (* consecutive same-name lines = one curried function *)
ElemPattern    ::= "_" | Identifier | Literal       (* nested patterns are not supported yet *)
(* A literal pattern (Number, "String", 'c', true, false, Duration, Time, Date) matches a value `v` exactly
   when `v == literal` under the language's own equality: `0` matches 0 and 0.0, never "0", '0', 10 or 20; a
   String is never a number. There is no substring matching. A value no arm matches is
   `[ERROR] match: no arm matches <value>`; a pattern that is none of the forms above (an
   expression `1 + 2`, a nested list) is an error naming the pattern when the match is built. *)

(* List Comprehensions (multi-generator, with filter guards) *)
Comprehension  ::= "[" Expr "|" Generator { "," ( Generator | Expr ) } "]"   (* at least one Generator: `[ e | guard ]` is an error; write `[ e | _ <- [0], guard ]` *)
Generator      ::= GenPattern "<-" Expr
GenPattern     ::= "_" | Identifier | "[" ElemPattern { "," ElemPattern } "]"   (* binds the element, or destructures a fixed-length one *)

(* Optics (Lenses) & Records *)
                 | "parse_record " { Identifier }
                 | "focus " Identifier
                 | "update " Identifier " " StringLiteral

(* Primitives *)
OSCommand      ::= Identifier { Identifier | StringLiteral | "%s" }
Constructor    ::= Identifier   (* an Identifier registered as a branch of a sum type *)
FuncName       ::= Identifier   (* an Identifier bound by a `def` *)
Expr           ::= (* infix arithmetic/relational/boolean expression, Application, Comprehension, or Literal *)
Application    ::= Atom { Atom }   (* juxtaposition: `f a b` applies f to a, b; binds tighter than any operator *)
Atom           ::= Identifier | Literal | StringLiteral | CharLiteral | "(" Expr ")" | "[" ... "]"
LambdaExpr     ::= "\\" Identifier { Identifier } "->" Expr   (* a Closure VALUE; body extends as far right as possible; parenthesise to pass as an argument *)
Literal        ::= Number | Duration | Time | Date | StringLiteral | CharLiteral | "[" [ Expr { "," Expr } ] "]"
Number         ::= Integer | Float | Complex | Ratio
Duration       ::= [ "+" | "-" ] Segment { Segment }   (* 5s  1h30m  250ms  2d  1w  1.5h ; a number IMMEDIATELY followed by a unit (`5 s` is an application) *)
Segment        ::= ( Digits | Digits "." Digits ) Unit   (* only the first segment may be fractional; a unit once; whole nanoseconds; int64 range *)
Unit           ::= "ns" | "us" | "µs" | "μs" | "ms" | "s" | "m" | "h" | "d" | "w"   (* d = 24 h and w = 7 d: fixed lengths, not calendar days *)
Time           ::= "@" Digits4 "-" Digits2 "-" Digits2 "T" Digits2 ":" Digits2 ":" Digits2 [ "." Digits ] ( "Z" | ( "+" | "-" ) Digits2 ":" Digits2 )   (* RFC 3339; the zone is required; years 1..9999 *)
Date           ::= "@" Digits4 "-" Digits2 "-" Digits2   (* ISO 8601, a real calendar day; years 1..9999 *)
LiteralToken   ::= Number | Duration | Time | Date | AnyText   (* the UNQUOTED value of `const`, `emit` and `-D`: a CANONICAL literal (prints back unchanged; Duration segments have no `+`, leading or trailing zeros) is that value, any other text, `007`, `007s`, `1.50`, a quoted const, is a String; `functions.LiteralToken` *)
AnyText        ::= ANY_CHAR { ANY_CHAR }   (* any other unquoted text is a String *)
Ratio          ::= Digits "r" Digits | Digits "." Digits "r"   (* 3r4 ; the exact decimal 0.75r ; 1r0 is a lex error, a bare 5r is an error *)
PrefixOp       ::= "-" | "!"   (* in operand position: `-x`, `!ok`; `!` only when not followed by "=" ; `f -3` (space before, none after) is a negative argument *)
BinaryOp       ::= "||" | "&&" | "==" | "!=" | "<" | "<=" | ">" | ">=" | "+" | "-" | "++" | "*" | "/" | "//" | "%" | "**"
                   (* tightest last; "**" is right-associative and binds tighter than a prefix operator *)
Identifier     ::= ALPHA { ALPHA | DIGIT | "_" }
StringLiteral  ::= "\"" { ANY_CHAR | CharEscape } "\""   (* the same escapes as a Char; an unknown escape is an error *)
CharLiteral    ::= "'" ( ANY_CHAR | CharEscape ) "'"   (* exactly one Unicode code point; a value of type Char *)
CharEscape     ::= "\\" ( "n" | "t" | "r" | "0" | "\\" | "'" | "\"" | "x" HEX HEX | "u" HEX HEX HEX HEX )
HEX            ::= DIGIT | "a".."f" | "A".."F"
Integer        ::= DIGIT { DIGIT }   (* unbounded: an int64 fast path and big integers, canonical after every operation *)
Float          ::= DIGIT { DIGIT } "." DIGIT { DIGIT }
               | DIGIT { DIGIT } [ "." DIGIT { DIGIT } ] ( "e" | "E" ) [ "+" | "-" ] DIGIT { DIGIT }   (* 1e3  2.5E-3 are Floats; `5em` / `2e` are not exponents *)
Complex        ::= ( Integer | Float ) "i"
Digits         ::= DIGIT { DIGIT }
Digits2        ::= DIGIT DIGIT
Digits4        ::= DIGIT DIGIT DIGIT DIGIT
ALPHA          ::= "A".."Z" | "a".."z"
DIGIT          ::= "0".."9"
```

### Optional type annotations and `gatwd --check` (gradual typing)

`def name : Type = body` carries an optional, purely advisory signature:
`def inc : Int -> Int = \x -> x + 1`, `def pairs : List(Int) -> Int -> Bool = …`.
`Type ::= Atom { "->" Type }` (right-associative), `Atom ::= Ident [ "(" Type { "," Type } ")" ] | "(" Type ")"`;
names declared by `type Alias = Int` resolve to their right-hand side. The annotation
never changes evaluation. `gatwd --check` reads it and, with local inference from
literals, operators, let bindings and comprehension generators, reports as
*warnings* (exit status unaffected) the definite contradictions it can see: wrong
argument or result types, too many arguments (to a def or a builtin), operators the
runtime matrix rejects for the operand types (`"a" - 1`, `[1] * 2`), unbound names in
expressions, and a `spawn IntVal: hello` seed that contradicts a numeric/Bool alias.
Anything it cannot determine is `Any`, which is consistent with every type — so valid
programs, and everything produced at runtime, are never flagged.

### Closures: lambdas as values, currying, lexical scope

`\x y -> x + y` is an *expression* that evaluates to a `functions.Closure` carrying
its parameter list and the name environment of its definition site (lexical scope).
Closures are **curried**: applying fewer arguments than the arity yields a closure
awaiting the rest (`add 1` ⇒ `\y -> 1 + y`), exactly the isomorphism
`(A×B → C) ≅ (A → B → C)`, and applying more arguments applies the result to the
surplus (`mk 5 10` where `mk = \k -> \x -> x + k`). A lambda-bodied `def` and a
`let`-bound lambda are the same kind of value, so they can be passed to
`map`/`filter`/`reject`/`fold` by name or as a partial application
(`map (add 10)`, `fold add 0`, `filter is_pos`). In pipeline position
`x |> f a b` applies `f a b x` — the incoming value is the LAST argument, the
Haskell data-last convention; a bare one-argument closure name maps over the incoming
collection. A `def` whose body is not a single lambda stays a pipeline macro.

### Function application in expressions (juxtaposition)

Inside an infix expression — a `fold`/`map`/`filter` reducer body, a comprehension
source/yield, or any `Expr` — a callable value is applied to arguments by
*juxtaposition*, exactly as in Haskell: `graph_add_node "Pod" id payload g`. The head
resolves to a `functions.DSLFunc` (a type-erased callable), each following *atom* is
an argument, and application binds tighter than every operator (`f x + g y` parses as
`(f x) + (g y)`). Nested application needs parentheses: `graph_add_node (hamt_get
"kind" r) (hamt_get "name" r) r g`.

The one subtlety is *plain list literals*, which gatwd serializes space-separated
(`[2 1]`, `[nginx web]`) and re-parses: inside a bare `[...]` adjacent atoms stay
**separate elements**, never an application. A comprehension (`[e | …]`) and a
parenthesised group `(…)` are full expression contexts where application applies
normally. This is what lets `fold`-based graph ingest — the accumulator is a graph
and the reducer calls `graph_add_node`/`graph_add_edges` to grow it — be written in
pure DSL rather than faked with a Go loop.

The generic, domain-free expression-level builtins currently registered are the
graph constructors (`graph_empty` value; `graph_add_node`, `graph_insert_node` (errors on a duplicate id), `graph_add_edge`,
`graph_add_inferred_edge`, `graph_add_edges`), the payload accessors (`hamt_get`,
`hamt_path`), and the containment predicate `subset_of sub super` (Record/HAMT
key-value containment or Set membership) — each the function form of the
identically-named pipeline step. A **bare sum-type constructor** used in an
expression (e.g. an edge color `Owns` from `type EdgeColor = Owns | …`) resolves to
its tag **string** `"Owns"`, so relation rules need not quote colors. These resolve
in every expression context (fold/map/filter reducer bodies, comprehension
sources/yields, bare expressions) via a low-priority lookup fallback, so a user
binding of the same name always wins.

The **operator-type matrix** is also available as expression functions (Workstream E),
all *data-last* and *curried*, preserving element types: `map filter reject fold
flatmap bimap`; sequences `length is_empty head tail nth reverse cons concat range
member to_list to_set`; Set `set_empty set_add set_remove set_contains set_union
set_intersect set_diff set_elements`; HAMT `hamt_empty hamt_from hamt_set hamt_del
hamt_has hamt_keys hamt_values hamt_pairs hamt_merge hamt_diff hamt_intersect`; Tree
`tree_leaf tree_node tree_add_child tree_value tree_children tree_size tree_depth
tree_flatten tree_member`; BST `bst_empty bst_from bst_insert bst_member`; Pair `pair
fst snd pair_swap`; Maybe/Either `just nothing is_just is_nothing maybe_default
maybe_fold either_left either_right is_left is_right either_default either_fold`;
Variant `variant variant_tag variant_payload is_variant` (`variant "Circle" 3` boxes a
declared constructor with its payload); primitives `min max abs negate compare`.
`map`/`filter`/`fold`/`bimap` dispatch on the carrier (List, Chunk, Set, HAMT values,
Tree, BST for fold, Pair second component, Maybe, Either-Right, Variant payload).
Errors are values: a wrong carrier, non-callable, out-of-range index or unhashable Set
element yields an `[ERROR]` result rather than a silent no-op. A call used as an
element of a *plain* list literal needs parentheses: `[(tree_leaf 2), (tree_leaf 3)]`.
