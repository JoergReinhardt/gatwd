# Prompt: add a graph relation, ingest pass or compliance rule (in the DSL)

**Task:** `<relation or rule, in domain terms, with an example input and expected output>`

**Read first:** `.agents/skills/dsl-authoring/SKILL.md`; `pkg/k8sgraph/lib/k8s/relations.gatwd`
and `cmd/gatwd-k8s-operator/rules/compliance.gatwd` (the style reference);
`pkg/k8sgraph/edition_test.go`, `cmd/gatwd-k8s-operator/main_test.go`.

**Steps**
1. **Write the fixture first**: a manifest/testdata file with a positive case, a near-miss, and
   an absent-field case. For a rule, add a case to the *violating* fixture
   (`testdata/broken.yaml`) so it is proven to fire.
2. **Write the expected output as a test** (`edgesOf`/`ordered` helpers; sort when order is not
   part of the contract).
3. **Author it in `.gatwd`, never Go.**
   - Relation: `def rel_x = fold (\_ g -> graph_add_edges [["Color", w, from, to] | … ] g) graph_empty`
     appended to the `k8s_relations` chain. Inferred evidence → `graph_add_inferred_edges`, weight < 1.
   - Rule: `def rule_x = \g -> [ ["Violation", resource] | … ]`, added to the `++` chain in
     `compliance`.
   - Shared sub-expressions → ordinary closures (`def k8s_ref = \kind ns name g -> […]`).
4. **Apply the traps**: nodes before edges; guard with `graph_has_node`; `hamt_has_path` *before*
   the dependent generator; parenthesise calls in list literals; the graph is the last argument.
5. **Extend the vocabulary** (`types.gatwd`: `NodeColor`/`EdgeColor`) so exhaustiveness and
   readers see the new name. Node colors are the manifest's real `kind`.
6. **If Go changes seem needed, stop**: the missing piece is probably a *generic* primitive.
   Add it to core with a non-domain DSL assertion (Mandate 3), then use it.
7. **Check performance** if the pass joins two node sets (quadratic risk): build an index in DSL
   (see the label-indexed `Selects`) and run `benchmarks/run_k8s_graph.sh`.

**Definition of done:** positive, near-miss and absent-field cases asserted; rule proven by the
violating fixture; live/offline parity test still green; vocabulary updated; gates green.

**Don't:** loop in Go · flag the compliant case only · let an empty result stand in for "works".
