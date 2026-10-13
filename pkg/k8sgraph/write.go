package k8sgraph

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/collections"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
	"github.com/joergreinhardt/gatwd/pkg/cli"
)

type (
	pipelineStep = evaluator.PipelineStep[functions.Either[error, any]]
	nodeOf       = dsltypes.ASTNode[functions.Either[error, any]]
	nodeSeq      = iter.Seq[nodeOf]
)

// This file is the WRITE half of the shim: it lifts the adapters.Applier (server-side apply,
// and later the rollout family) to DSL steps. Like the read half it holds no logic: it names
// the steps, wraps every one in cli.GatedStep so the capability policy decides whether a
// script may write at all, and turns the adapter's WriteRecord into the list the DSL sees.
//
// Three layers of defence, all denied by default:
//
//  1. the capability policy: cli.CapClusterApply / cli.CapClusterRollout, which not even the
//     trusted command-line script has until the user passes --allow-cluster-apply /
//     --allow-cluster-rollout;
//  2. server dry run: every request carries DryRun=All unless --dry-run=false is given;
//  3. RBAC: the identity the client runs as (examples/k8s_howto/deploy-write-role.yaml).
//
// There is no step that deletes or prunes, and the adapter cannot express one.

// WriteFlags are the command-line knobs of the write steps. Bind declares them on a flag set;
// a host that has its own flag syntax (the operator Appliance) fills the fields itself and
// hands the struct to WithWriteFlags.
type WriteFlags struct {
	// DryRun is true (the default) for a server dry run: nothing is persisted. --dry-run=false
	// is the explicit switch that makes a write real.
	DryRun bool
	// ForceConflicts takes over fields another field manager owns (--force-conflicts).
	ForceConflicts bool
	// FieldManager is the server-side apply field manager (--field-manager); "" means gatwd.
	FieldManager string
	// Namespace is the namespace of a namespaced object that names none (--apply-namespace).
	Namespace string
	// Out, when set, is the offline twin: k8s_apply renders each object to a YAML file in this
	// directory instead of talking to a cluster (--out DIR). Naming the directory is the
	// opt-in, so --dry-run does not apply to it; it still needs the fs capability.
	Out string
}

// DefaultWriteFlags is the safe starting point: a dry run, no forced conflicts.
func DefaultWriteFlags() *WriteFlags { return &WriteFlags{DryRun: true} }

// Bind declares the flags on f, bound to w, with the defaults of w.
func (w *WriteFlags) Bind(f *pflag.FlagSet) {
	f.BoolVar(&w.DryRun, "dry-run", w.DryRun, "server dry run for the write steps: validate and report, persist nothing; --dry-run=false makes the writes real")
	f.BoolVar(&w.ForceConflicts, "force-conflicts", w.ForceConflicts, "let k8s_apply take over fields another field manager owns")
	f.StringVar(&w.FieldManager, "field-manager", w.FieldManager, "server-side apply field manager `NAME` (default \"gatwd\")")
	f.StringVar(&w.Out, "out", w.Out, "render the objects of k8s_apply as YAML files (one per object) into `DIR` instead of writing to a cluster; needs no cluster")
	f.StringVar(&w.Namespace, "apply-namespace", w.Namespace, "namespace `NS` for a namespaced object that names none (default: such an object is an error)")
}

// Settings reads the flags as the adapter's WriteSettings; called for every object written.
func (w *WriteFlags) Settings() adapters.WriteSettings {
	return adapters.WriteSettings{
		Persist:        !w.DryRun,
		FieldManager:   w.FieldManager,
		ForceConflicts: w.ForceConflicts,
		Namespace:      w.Namespace,
	}
}

// WithWriteFlags makes ApplyEdition read the write settings from w instead of declaring the
// flags on the edition itself: for a host (the operator Appliance) that owns its command line
// and its capability flags.
func WithWriteFlags(w *WriteFlags) Option { return func(c *editionConfig) { c.write = w } }

// WithKindResolver replaces the kind resolver of the write steps (default: the built-in kind
// table, then discovery against the kubeconfig cluster).
func WithKindResolver(r adapters.ResourceResolver) Option {
	return func(c *editionConfig) { c.resolver = r }
}

// defaultResolver tries the built-in kind table, then asks the cluster (lazily: a kubeconfig
// is read only for a kind the table does not know).
func defaultResolver() adapters.ResourceResolver {
	return adapters.ChainResolvers(adapters.BuiltinKindResolver(),
		adapters.DiscoveryResolver(func() (discovery.DiscoveryInterface, error) { return DiscoveryClientFromKubeconfig() }))
}

// writeTimeout bounds one object's round trips to the API server.
const writeTimeout = 60 * time.Second

// recordOp is one write operation on one object: it returns the record of what it did.
type recordOp func(ctx context.Context, interp *cli.Interpreter, arg string, v any) (adapters.WriteRecord, error)

// recordStep builds a step that runs op on every object of the stream and yields the record
// as the DSL list [verb, id, outcome]. A value that is a LIST is a batch: op runs on each
// element in list order and yields one record each, so `k8s_graph |> k8s_apply_objects |>
// k8s_apply` applies in the order the DSL computed. An error is an error node for that object
// (Mandate 6) and the stream goes on with the next one; an error node of the input passes
// through. recordStep creates the step but does not gate it: the caller MUST wrap the result
// in cli.GatedStep (guarded by TestEveryWriteStepIsGated).
func recordStep(name string, op recordOp) cli.StepHandler {
	return func(stepStr string, interp *cli.Interpreter) (pipelineStep, bool) {
		trimmed := strings.TrimSpace(stepStr)
		if trimmed != name && !strings.HasPrefix(trimmed, name+" ") {
			return nil, false
		}
		arg := strings.TrimSpace(strings.TrimPrefix(trimmed, name))
		return func(in nodeSeq) nodeSeq {
			return func(yield func(nodeOf) bool) {
				one := func(src nodeOf, v any) bool {
					ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
					defer cancel()
					rec, err := op(ctx, interp, arg, v)
					if err != nil {
						return yield(dsltypes.NewASTNode(functions.Left[error, any](err), src.DSLType))
					}
					return yield(dsltypes.NewASTNode(functions.Right[error, any](recordList(rec)), src.DSLType))
				}
				for node := range in {
					if node.Value.IsLeft() {
						if !yield(node) {
							return
						}
						continue
					}
					v, _ := node.Value.UnwrapRight()
					if batch, ok := asBatch(v); ok {
						for _, item := range batch {
							if !one(node, item) {
								return
							}
						}
						continue
					}
					if !one(node, v) {
						return
					}
				}
			}
		}, true
	}
}

// asBatch reports whether v is a list of objects (a DSL List or a Go slice), and its elements.
func asBatch(v any) ([]any, bool) {
	switch t := v.(type) {
	case *collections.List[any]:
		return t.Elements(), true
	case []any:
		return t, true
	}
	return nil, false
}

// recordList is the DSL value of a record: the list [verb, id, outcome].
func recordList(r adapters.WriteRecord) any {
	return collections.Cons[any](r.Verb, collections.Cons[any](r.ID, collections.Cons[any](r.Outcome, collections.Nil[any]())))
}

// ApplyStep is the `k8s_apply` step: server-side apply of every object of the stream (or of a
// list of objects, in list order), one [apply, Kind/ns/name, outcome] record each. It is denied
// unless the policy grants cli.CapClusterApply, and a dry run unless the settings say
// otherwise (see WriteFlags). With out() naming a directory it renders the objects there
// instead (the offline twin, see adapters.Applier.RenderTo), which additionally needs the fs
// capability. a.Client may be nil (no cluster): applying then fails loudly per object.
func ApplyStep(keyword string, a *adapters.Applier, out func() string) cli.StepHandler {
	return cli.GatedStep(cli.CapClusterApply, recordStep(keyword, func(ctx context.Context, interp *cli.Interpreter, arg string, v any) (adapters.WriteRecord, error) {
		if strings.TrimSpace(arg) != "" {
			return adapters.WriteRecord{}, fmt.Errorf("%s takes no argument (got %q); set the behavior with --dry-run, --force-conflicts, --field-manager, --out", keyword, arg)
		}
		if dir := out(); dir != "" {
			if err := interp.RequireCapability(cli.CapFS, keyword, dir); err != nil {
				return adapters.WriteRecord{}, err
			}
			return a.RenderTo(dir, v)
		}
		return a.Apply(ctx, v)
	}))
}

// newApplier assembles the adapter the write steps share: the client, the kind resolver
// (default: built-ins, then discovery) and the flag-backed settings.
func newApplier(client dynamic.Interface, w *WriteFlags, r adapters.ResourceResolver) *adapters.Applier {
	if r == nil {
		r = defaultResolver()
	}
	return &adapters.Applier{Client: client, Resolver: r, Settings: w.Settings}
}
