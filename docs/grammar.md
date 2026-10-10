# gatwd Formal Grammar (EBNF)

This is the formal Extended Backus-Naur Form (EBNF) grammar for the `gatwd` Domain Specific Language.

```ebnf
Program        ::= { Statement }
Statement      ::= TypeDef | Pipeline | Comment
Comment        ::= "--" { ANY_CHAR } "\n"

(* Type Definitions (Algebraic Data Types) *)
TypeDef        ::= "type " Identifier "=" (SumType | RecordType | TupleType)
SumType        ::= Identifier { "|" Identifier }
RecordType     ::= "{" FieldDef { "," FieldDef } "}"
TupleType      ::= "(" Identifier { "," Identifier } ")"
FieldDef       ::= Identifier ":" Identifier

(* Pipelines (Graph Traversal & Evaluation) *)
Pipeline       ::= SpawnNode { ( "|>" | "||>" ) PipelineStep }
SpawnNode      ::= "spawn " Identifier ":" Identifier

(* Core Mechanics & Transducers *)
PipelineStep   ::= OSCommand 
                 | "take " Integer
                 | "drop " Integer
                 | "fold " ("count" | "sum")
                 | "distinct"
                 | "memoize " OSCommand

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
MatchRule      ::= Identifier "->" OSCommand

(* Optics (Lenses) & Records *)
                 | "parse_record " { Identifier }
                 | "focus " Identifier
                 | "update " Identifier " " StringLiteral

(* Primitives *)
OSCommand      ::= Identifier { Identifier | StringLiteral | "%s" }
Identifier     ::= ALPHA { ALPHA | DIGIT | "_" }
StringLiteral  ::= "\"" { ANY_CHAR } "\""
Integer        ::= DIGIT { DIGIT }
ALPHA          ::= "A".."Z" | "a".."z"
DIGIT          ::= "0".."9"
```
