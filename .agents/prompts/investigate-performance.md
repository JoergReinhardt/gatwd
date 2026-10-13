# Prompt: investigate and fix a performance problem

**Task:** `<what is slow, at what size, how it scales>`

**Read first:** `docs/development_playbook.md` §7, `benchmarks/README.md`,
`pkg/k8sgraph/bench_test.go`, `benchmarks/run_k8s_graph.sh`.

**Steps**
1. **Reproduce with a scalable synthetic workload** and a benchmark (`go test -bench`, several
   sizes). Record the numbers and the growth ratio; a quadratic cost shows as ≈4× per 2×.
2. **Profile per stage** — time each pass/step separately (or `pprof`) before changing anything.
   Name the dominant stage with its share.
3. **Pin semantics**: write the test(s) that define correct behavior including edge cases
   (empty selector, missing field, unlabeled item) *before* optimising.
4. **Fix at the owning layer.** In DSL code prefer an index (`hamt_set`/`hamt_get_or` fold) to a
   Go special case; in the interpreter prefer removing repeated parsing to adding caches (the
   def-step cache gave ~1% — measure before building more).
5. **Re-measure** the same benchmark; confirm the growth ratio changed, not just the constant.
6. **Run the gates and the script regression diff**; run `-race` if parallel steps are involved.
7. **Record honestly** in `benchmarks/README.md`, including a negative result; state that numbers
   are single runs on one machine.

**Definition of done:** before/after table, semantics tests, gates green, README updated.

**Don't:** optimise the stage you guessed · change behavior and speed in one step · cache without
proving the cached parse is identical to a fresh one.
