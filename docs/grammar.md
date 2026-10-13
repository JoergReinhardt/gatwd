# gatwd Formal Grammar (EBNF)

This is the formal Extended Backus-Naur Form (EBNF) grammar for the `gatwd` Domain Specific Language.

```ebnf
Program        ::= { Statement }
Statement      ::= Import | TypeDef | FunctionDef | LetBinding | Pipeline | Comment
Comment        ::= "--" { ANY_CHAR } "\n"

(* Modularity & Reusability *)
Import         ::= "import " StringLiteral
FunctionDef    ::= "def " Identifier "=" Pipeline
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
                 | LetBinding                      (* let x = Pipeline in Pipeline, as a step *)

FoldOp         ::= "count" | "sum" | "product" | "min" | "max" | "concat" | "any" | "all"
Lambda         ::= [ "(" ] "\\" Identifier "->" Pipeline [ ")" ]

(* Error Handling & Compliance *)
                 | "expect " StringLiteral " " StringLiteral
                 | "filter " StringLiteral

(* Arrow Routing (Kleisli & Applicative) *)
                 | "(" OSCommand " &&& " OSCommand ")"
                 | "[" OSCommand " *** " OSCommand "]"
                 | "zip " OSCommand
                 | "format_tuple " StringLiteral

(* Pattern Matching (Catamorphisms) *)
                 | "match {" MatchRule { "|" MatchRule } "}"
MatchRule      ::= Pattern "->" Pipeline
Pattern        ::= Constructor                      (* registered sum-type branch: discriminates *)
                 | "_"                              (* wildcard catch-all *)
                 | Identifier                        (* fresh variable: binds any value *)
                 | "[]"                             (* empty list *)
                 | "[" Identifier ", ..." Identifier "]"   (* head / tail *)
                 | "[" Pattern { "," Pattern } "]"  (* fixed-length list/tuple, literals or binders *)
                 | Literal

(* List Comprehensions (multi-generator, with filter guards) *)
Comprehension  ::= "[" Expr "|" Generator { "," ( Generator | Expr ) } "]"
Generator      ::= Pattern "<-" Expr

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
Atom           ::= Identifier | Literal | StringLiteral | "(" Expr ")" | "[" ... "]"
Literal        ::= Integer | StringLiteral | "[" [ Expr { "," Expr } ] "]"
Identifier     ::= ALPHA { ALPHA | DIGIT | "_" }
StringLiteral  ::= "\"" { ANY_CHAR } "\""
Integer        ::= DIGIT { DIGIT }
ALPHA          ::= "A".."Z" | "a".."z"
DIGIT          ::= "0".."9"
```

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
graph constructors (`graph_empty` value; `graph_add_node`, `graph_add_edge`,
`graph_add_inferred_edge`, `graph_add_edges`), the payload accessors (`hamt_get`,
`hamt_path`), and the containment predicate `subset_of sub super` (Record/HAMT
key-value containment or Set membership) — each the function form of the
identically-named pipeline step. A **bare sum-type constructor** used in an
expression (e.g. an edge color `Owns` from `type EdgeColor = Owns | …`) resolves to
its tag **string** `"Owns"`, so relation rules need not quote colors. These resolve
in every expression context (fold/map/filter reducer bodies, comprehension
sources/yields, bare expressions) via a low-priority lookup fallback, so a user
binding of the same name always wins.
