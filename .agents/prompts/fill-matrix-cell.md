# Prompt: fill an operator-type matrix cell (or row/column)

**Task:** `<operation>` exists for `<types>` but not for `<missing types>` (or: add `<new type>`).

**Read first:** `STYLEGUIDE.md` §7, `AGENTS.md` Mandate 2, `.agents/skills/recursive-dogfood`,
`pkg/cli/dslfuncs_matrix.go` (header lists every cell), `pkg/cli/valueops.go`.

**Steps**
1. **Draw the matrix** for the operation across List, Chunk, Set, HAMT, Tree, BST, Graph,
   Pair, Either, Maybe, Variant, primitives. Mark each cell: present, missing-but-valid,
   mathematically meaningless (write the reason).
2. For every *missing-but-valid* cell, in this order: Go method on the type (+ structural
   tests with diverse element types) → operator overload in `functions/operators.go` if infix →
   `evaluator` step if a stream op → DSL **step form** → DSL **function form** (`defn`, data-last,
   curried) → carrier case in `valueops.go` → DSL assertions → `Example`.
3. **Mind the capacity/identity traps** found last time: `Chunk` has fixed capacity (`Chunk.Bind`
   truncated past it — size results to the total); `Set` from the `to_set` step is `Set[string]`
   while function forms normalise to `Set[any]`; Go methods cannot add type parameters, so
   `Map` is `func(T) T` on collections.
4. **Dogfood**: can older code now use the new cell?

**Definition of done:** every valid cell filled through every layer; every skipped cell has a
written mathematical reason in the code comment or phase entry; gates green.

**Don't:** implement `List` and leave `Chunk`/`Set`/`HAMT` for later · skip because it is tedious.
