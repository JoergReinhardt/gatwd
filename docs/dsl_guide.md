# gatwd DSL Guide

`gatwd` is a purely functional, Turing-complete Scripting Language for declarative infrastructure crawling, evaluation, and mutation. It heavily borrows syntax and philosophy from Haskell, Elm, and F#.

## 1. Modules & Reusability

In `gatwd`, you don't write monolithic scripts. You define **Types** and **Functions** (`def`) in library modules, and bring them into your runtime using `import`.

**types.gatwd**
```haskell
-- Sum Types (Dynamic Enums)
type OS = Linux | Windows | Darwin

-- Record & Collection Types
type ServerConfig = { IP: String, Port: Int }
type Cluster = Set(ServerConfig)
```

**main.gatwd**
```haskell
import "types.gatwd"

-- Functions (Aliases) encapsulate pipeline logic
def audit_linux = uname -a |> expect "Linux" "CRITICAL: Not Linux"
def extract_ip = parse_record IP Port |> focus IP

spawn OS: Linux |> audit_linux |> extract_ip
```

## 2. Local State (`let ... in`)

While pipelines (`|>`) flow data sequentially, you often need to store intermediate state. Use `let` bindings to capture terminal nodes and substitute them into subsequent branches.

```haskell
let master_ip = spawn Server: Master |> focus IP in
spawn Server: Worker |> echo Routing to master_ip!
```

## 3. Lambdas & Transducers

`gatwd` utilizes lazy `iter.Seq` trampolines, meaning graphs can be infinite.
* `take N` / `drop N`: Bounding infinite streams.
* `distinct`: Drops duplicates (critical for cyclic graphs).
* `fold count`: Catamorphism that collapses streams to a scalar.
* `map (\x -> ...)`: Anonymous functions for inline mapping.

```haskell
def list_files = ls -l
def count_gatwd = filter "gatwd" |> fold count

spawn Directory: . |> list_files |> map (\x -> echo Processing x...) |> count_gatwd
```

## 4. Arrow Routing (Zipping & Splitting)

You can mathematically route streams without breaking the functional pipeline:
* `(A &&& B)` (Kleisli FanOut): Routes a node into TWO independent functions concurrently, yielding a Tuple `(Res1, Res2)`.
* `[A *** B]` (Applicative Split): Takes a Tuple, and routes the Left side to Function A, and Right side to Function B concurrently.

```haskell
def get_time = date
def get_user = whoami

spawn OS: Linux |> (get_time &&& get_user) |> [echo Time_%s *** echo User_%s] |> format_tuple "(%s, %s)"
```

## 5. Optics (Lenses) & Zippers (Comonads)

`gatwd` natively parses unstructured text into structural collections (Records, Trees), allowing you to manipulate them mathematically.

**Optics (Records)**
```haskell
def parse_file = parse_record Perms Links Owner Name
def secure_file = update Owner "root" |> focus Name

spawn Directory: /etc |> ls -l |> parse_file |> secure_file
```

**Zippers (Trees)**
```haskell
-- Build a Rose Tree from a stream, wrap it in a Zipper context, and navigate O(1)
spawn Pod: Kube |> echo Child_1 |> echo Child_2 |> build_tree "Root" |> into_zipper |> down |> extract
```
