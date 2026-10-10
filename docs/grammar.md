# gatwd DSL Formal Grammar & Syntax

The `gatwd` DSL is designed to expose the full, raw power of the underlying Go Host Engine. This means providing syntax not just for simple graph traversal, but for all the advanced Category Theory constructs we built: Optics (Lenses), Comonads (Zippers), Arrows (Split/Fan-out), and Transducers (Stream combinations).

## Overcoming Haskell's Annoyances

1. **Cryptic Operators (`>>=`, `<*>`, `$`):**
   Instead of Haskell's abstract monadic operators, we use the highly readable **Pipe Operator (`|>`)**. Because our entire engine runs on `iter.Seq` (a lazy stream), `|>` perfectly represents binding a stream of nodes into the next Adapter.
2. **The "Layout Rule" (Whitespace Hell):**
   We use a forgiving layout: newlines signify the end of an expression, but if a line ends with an operator (like `|>`), the parser inherently knows the expression continues on the next line.

---

## Elaborate Example Script

This script demonstrates every advanced mechanic available in the engine.

```haskell
-- 1. Algebraic Data Types (Dynamic BitFlags)
type Server = Linux | Windows
type KubeNode = { Name: String, Status: String }

-- 2. Zippers (Comonadic Navigation)
-- Use 'up', 'down', 'left', 'right' to move the Comonadic cursor around an AST
navigateGraph = 
    spawn KubeNode "node-01" 
    |> down                          -- Move cursor to the child (Pods)
    |> right                         -- Move cursor to sibling Pod
    |> extract                       -- Comonad Extract: Get the focused value

-- 3. Optics (Lenses)
-- Use 'focus' (Lens Get) and 'update' (Lens Set) for deep immutability
patchIP = 
    spawn Server "db-server"
    |> update { Network.IP: "10.0.0.1" }  -- Returns a BRAND NEW immutable graph

-- 4. Arrows (Split & Fan-out)
-- The '&&&' operator routes a single node into two different Adapters concurrently,
-- returning a stream of Pairs (Zip)!
monitorServer = 
    spawn Server "prod-server"
    |> (getLogs &&& getMetrics)      -- Kleisli Arrow Fan-Out!

-- 5. Transducers & Stream Bounding
-- Use 'take', 'drop', and 'zip' on lazy infinite streams
sampleNodes = 
    spawn Datacenter "us-east"
    |> getServers
    |> filter (\n -> n.Status == "Active")
    |> take 5                        -- Prevents infinite loops!

-- 6. Parallel Execution (ParBind)
-- The ||> operator invokes the ParBind worker pool
auditServers =
    spawn Datacenter "eu-west"
    |> getServers
    ||> runSecurityScan              -- Evaluated concurrently across multiple goroutines!
```

---

## Formal EBNF (Extended Backus-Naur Form)

This is the comprehensive mathematical blueprint mapping to our Parser Combinators.

```ebnf
Program         ::= Statement*

Statement       ::= TypeDef | Assignment | Expression

-- Algebraic Data Types
TypeDef         ::= "type" Identifier "=" TypeExpr
TypeExpr        ::= SumType | RecordType
SumType         ::= Identifier ("|" Identifier)*
RecordType      ::= "{" FieldDef ("," FieldDef)* "}"
FieldDef        ::= Identifier ":" Identifier

-- Core Execution
Assignment      ::= Identifier "=" Expression
Expression      ::= Pipeline
Pipeline        ::= ArrowExpr (PipeOp ArrowExpr)*
PipeOp          ::= "|>" | "||>"

-- Arrows (Combinatorial Logic)
ArrowExpr       ::= Application
                  | "(" Application "&&&" Application ")"   -- Fan-out
                  | "(" Application "***" Application ")"   -- Split

-- Function Application & Built-in Constructs
Application     ::= Term+
Term            ::= Identifier | Literal | Lambda | "(" Expression ")"
                  | ZipperOp | OpticOp | StreamOp

-- Advanced Category Theory Built-ins
ZipperOp        ::= "up" | "down" | "left" | "right" | "extract"
OpticOp         ::= "focus" Identifier | "update" RecordLiteral
StreamOp        ::= "take" NumberLiteral | "drop" NumberLiteral | "zip"

-- Primitives
Lambda          ::= "\\" Identifier "->" Expression
RecordLiteral   ::= "{" Identifier ":" Literal "}"
Identifier      ::= [a-zA-Z_] [a-zA-Z0-9_]*
Literal         ::= StringLiteral | NumberLiteral
StringLiteral   ::= '"' [^"]* '"'
NumberLiteral   ::= [0-9]+
```
