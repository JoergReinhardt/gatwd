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
Expr           ::= (* infix arithmetic/relational/boolean expression, Comprehension, or Literal *)
Literal        ::= Integer | StringLiteral | "[" [ Expr { "," Expr } ] "]"
Identifier     ::= ALPHA { ALPHA | DIGIT | "_" }
StringLiteral  ::= "\"" { ANY_CHAR } "\""
Integer        ::= DIGIT { DIGIT }
ALPHA          ::= "A".."Z" | "a".."z"
DIGIT          ::= "0".."9"
```
