# gatwd DSL Formal Grammar & Syntax

The `gatwd` DSL is heavily inspired by Haskell, ML, and Elm. However, we have deliberately designed it to avoid the most common annoyances of pure functional languages.

## Overcoming Haskell's Annoyances

1. **Cryptic Operators (`>>=`, `<*>`, `$`):**
   Instead of Haskell's abstract monadic operators, we use the highly readable **Pipe Operator (`|>`)** popularized by F# and Elixir. Because our entire engine runs on `iter.Seq` (a lazy stream), `|>` perfectly represents mapping/binding a stream of nodes into the next Adapter.
   *For Parallel Execution (our `ParBind`), we use the double-pipe (`||>`).*
2. **The "Layout Rule" (Whitespace Hell):**
   Haskell's strict indentation rules are notoriously difficult to parse and write. We will use a forgiving layout: newlines signify the end of an expression, but if a line ends with an operator (like `|>`), the parser knows the expression continues on the next line.
3. **Separated Type Signatures:**
   Haskell forces you to write the type signature on a separate line from the function. We allow inline type annotations where necessary, though our Hindley-Milner engine will infer most types automatically.

---

## Example Script

```haskell
-- 1. Algebraic Data Types (Dynamic BitFlags)
type Server = Linux | Windows
type Client = iOS | Android
type Endpoint = Server | Client

-- Record Types (Product Types)
type KubeNode = { Name: String, Status: String }

-- 2. Graph Traversal (The IO Monad / Trampoline)
-- 'crawlCluster' is a lazy sequence evaluated on the fly
crawlCluster = 
    spawn KubernetesCluster "prod-cluster-01"
    |> getNodes                               -- Adapter: Cluster -> Seq[Node]
    |> filter (\n -> n.Status == "Ready")     -- Lambda function
    |> getPods                                -- Adapter: Node -> Seq[Pod]
    |> select Image                           -- Optic Lens extraction

-- 3. Parallel Traversal
-- The ||> operator invokes the worker pool (ParBind)
auditServers =
    spawn Datacenter "us-east"
    |> getServers
    ||> runSecurityScan                       -- Executes concurrently!
```

---

## Formal EBNF (Extended Backus-Naur Form)

This is the mathematical blueprint we will use to build our Parser Combinators.

```ebnf
Program         ::= Statement*

Statement       ::= TypeDef | Assignment | Expression

-- Type Definitions (Phase 2 BitFlags)
TypeDef         ::= "type" Identifier "=" TypeExpr
TypeExpr        ::= SumType | RecordType
SumType         ::= Identifier ("|" Identifier)*
RecordType      ::= "{" FieldDef ("," FieldDef)* "}"
FieldDef        ::= Identifier ":" Identifier

-- Expressions & Pipelining (Phase 3 & 6 Evaluator)
Assignment      ::= Identifier "=" Expression
Expression      ::= Pipeline
Pipeline        ::= Application (PipeOp Application)*
PipeOp          ::= "|>" | "||>"

-- Function Application & Lambdas
Application     ::= Term+
Term            ::= Identifier | Literal | Lambda | "(" Expression ")"
Lambda          ::= "\\" Identifier "->" Expression

-- Primitives
Identifier      ::= [a-zA-Z_] [a-zA-Z0-9_]*
Literal         ::= StringLiteral | NumberLiteral
StringLiteral   ::= '"' [^"]* '"'
NumberLiteral   ::= [0-9]+
```
