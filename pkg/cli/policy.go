package cli

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"sync/atomic"

	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/functions"
)

// This file is the execution policy of the interpreter (hardening plan W7, decision D8).
//
// # The problem
//
// A script can do more than compute. Written in the pipeline language it can start an OS
// command (any word that is no step is run as a program), read and write files, import other
// files, and open sockets. That is exactly what a trusted command-line user wants, and
// exactly what must not be available to a script that arrived some other way: a ConfigMap
// override mounted into an operator, or a script submitted to an API.
//
// # The model (capability-based security, in one paragraph)
//
// Instead of asking "who is the author?" at every step, the interpreter asks "does the
// current Policy grant this capability?" A Capability names a class of side effect (exec,
// fs, net, env); a Policy is the set of capabilities granted. A step that needs a capability
// the policy denies does not run: it yields an error value (`[ERROR] step 'printf' requires
// the exec capability (policy denies it)`), on the same Left channel as every other failure
// (Mandate 6) — never an empty result, never a silent skip.
//
// # What is gated, and what is not
//
// The policy gates only what script TEXT can reach: the unknown-step OS-command fallback
// and the `memoize` and `zip` steps that take a command, the file steps (`walk_fs`,
// `write_file`, `append_file`, `scan_file`, `scan_words`) and file `import`, and the network
// steps (`dial_tcp`, `listen_tcp`). Steps an Edition wires itself (`RegisterStep`,
// `AdapterStep`: the Kubernetes crawl, `manifests`) are the author's decision, made in Go,
// and stay allowed. The one builtin that reads the environment, `env_get NAME`, is gated
// by AllowEnv; an Edition step that does the same consults Interpreter.RequireCapability.
//
// # The two capabilities even a trusted script does not get
//
// CapClusterApply and CapClusterRollout name WRITES to a remote system of record (an
// Edition's steps that create or change objects, and that restart, roll back, pause or
// scale a workload). Everything above can hurt the machine the script runs on; these can
// hurt a shared production system, often by a typo, and a script is easy to run by habit.
// So they are denied BY DEFAULT EVERYWHERE: TrustedPolicy does not grant them, a script the
// user names on the command line needs the explicit opt-in flag of the Edition
// (Edition.WithCapabilityFlag, e.g. --allow-cluster-apply and --allow-cluster-rollout), and an
// Appliance has the same two flags. They gate only the steps an Edition registers behind
// GatedStep; core itself has no step that needs them.
//
// # Who is trusted
//
// The script a person names on the command line (and `-e`, preludes, the REPL) is trusted:
// NewInterpreter starts with TrustedPolicy. Anything that did not come from the command line
// — an Appliance override, a RunOptions.Source — starts from DenyAll, and nothing inside a
// script can widen the policy: it is set by Go code (SetPolicy, RunOptions.Policy, the
// Appliance's WithPolicy) and read by the steps; no DSL statement names it.

// Capability names one class of side effect a script may be allowed to perform.
type Capability string

// The capabilities a Policy can grant.
const (
	// CapExec is running operating-system commands: the unknown-step fallback, `memoize CMD`, `zip CMD`.
	CapExec Capability = "exec"
	// CapFS is touching the file system: `walk_fs`, `write_file`, `append_file`, `scan_file`,
	// `scan_words`, and `import` of a file from disk (an embedded import is not file access).
	CapFS Capability = "fs"
	// CapNet is opening sockets: `dial_tcp`, `listen_tcp`.
	CapNet Capability = "net"
	// CapEnv is reading the process environment: `env_get NAME`.
	CapEnv Capability = "env"
	// CapClusterApply is creating or updating objects in a remote cluster (a declarative
	// apply). It is NOT part of TrustedPolicy: it needs an explicit opt-in, even for the
	// user's own command-line script.
	CapClusterApply Capability = "cluster-apply"
	// CapClusterRollout is operating on running workloads of a remote cluster: restart,
	// undo, pause, resume and scale (the rollout family; reading the rollout status is a
	// read and needs no capability). It is NOT part of TrustedPolicy either.
	CapClusterRollout Capability = "cluster-rollout"
)

// allCapabilities lists every capability, in the order Policy.String prints them.
var allCapabilities = []Capability{CapExec, CapFS, CapNet, CapEnv, CapClusterApply, CapClusterRollout}

// Policy is the set of capabilities a script is granted. The zero value grants nothing, so a
// forgotten field fails closed.
type Policy struct {
	// AllowExec grants CapExec.
	AllowExec bool
	// AllowFS grants CapFS.
	AllowFS bool
	// AllowNet grants CapNet.
	AllowNet bool
	// AllowEnv grants CapEnv.
	AllowEnv bool
	// AllowClusterApply grants CapClusterApply. TrustedPolicy leaves it false.
	AllowClusterApply bool
	// AllowClusterRollout grants CapClusterRollout. TrustedPolicy leaves it false.
	AllowClusterRollout bool
}

// TrustedPolicy grants every capability that acts on the machine the script runs on (exec,
// fs, net, env): the policy of a script the user named on the command line (and of
// NewInterpreter). It deliberately does NOT grant CapClusterApply or CapClusterRollout, the
// capabilities that write to a remote system; those need an explicit opt-in (see Grant and
// Edition.WithCapabilityFlag) even here.
func TrustedPolicy() Policy {
	return Policy{AllowExec: true, AllowFS: true, AllowNet: true, AllowEnv: true}
}

// DenyAll grants nothing: the default for a script that did not come from the command line.
func DenyAll() Policy { return Policy{} }

// Allows reports whether the policy grants c. An unknown capability is never granted.
func (p Policy) Allows(c Capability) bool {
	switch c {
	case CapExec:
		return p.AllowExec
	case CapFS:
		return p.AllowFS
	case CapNet:
		return p.AllowNet
	case CapEnv:
		return p.AllowEnv
	case CapClusterApply:
		return p.AllowClusterApply
	case CapClusterRollout:
		return p.AllowClusterRollout
	}
	return false
}

// Grant returns the policy with c added; an unknown capability is returned unchanged (it is
// never granted). It is how a command-line flag opts in to a capability TrustedPolicy
// withholds.
func (p Policy) Grant(c Capability) Policy {
	switch c {
	case CapExec:
		p.AllowExec = true
	case CapFS:
		p.AllowFS = true
	case CapNet:
		p.AllowNet = true
	case CapEnv:
		p.AllowEnv = true
	case CapClusterApply:
		p.AllowClusterApply = true
	case CapClusterRollout:
		p.AllowClusterRollout = true
	}
	return p
}

// String lists the granted capabilities, `exec,fs`, or `none`.
func (p Policy) String() string {
	var granted []string
	for _, c := range allCapabilities {
		if p.Allows(c) {
			granted = append(granted, string(c))
		}
	}
	if len(granted) == 0 {
		return "none"
	}
	return strings.Join(granted, ",")
}

// CapabilityError is the failure of a step that needs a capability the policy denies. It is
// a value on the Left channel like any other error; FailureOf reads its fields.
type CapabilityError struct {
	// Capability is the capability the step needs.
	Capability Capability
	// Step is the step (or statement keyword) that needed it: `printf`, `scan_file`, `import`.
	Step string
	// Detail is the subject when the step alone does not say it, e.g. the imported path.
	Detail string
}

// Error is the one-line message: `step 'printf' requires the exec capability (policy denies it)`.
func (e *CapabilityError) Error() string {
	subject := fmt.Sprintf("step '%s'", e.Step)
	if e.Step == "import" {
		subject = fmt.Sprintf("import %q", e.Detail)
	}
	return fmt.Sprintf("%s requires the %s capability (policy denies it)", subject, e.Capability)
}

// Require returns a *CapabilityError when the policy denies c, naming step (and detail).
func (p Policy) Require(c Capability, step, detail string) error {
	if p.Allows(c) {
		return nil
	}
	return &CapabilityError{Capability: c, Step: step, Detail: detail}
}

// execState is the execution environment one interpreter and all its forks share: the
// policy in force and the context that bounds the run. It is shared by pointer because
// compiled pipelines are cached and forks run `||>` in parallel; a step reads the state when
// the stream RUNS, so a policy or context installed after a step was compiled still applies
// and a fork can never hold a more permissive copy.
type execState struct {
	policy atomic.Pointer[Policy]
	ctx    atomic.Pointer[ctxBox]
	// budget, when set, bounds the bytes EvaluateToValues collects (Appliance.WithMaxOutput).
	budget atomic.Pointer[outputBudget]
	// calls counts closure calls across the run, to sample the context cheaply.
	calls atomic.Uint32
}

type ctxBox struct{ ctx context.Context }

func newExecState() *execState {
	s := &execState{}
	p := TrustedPolicy()
	s.policy.Store(&p)
	return s
}

// current returns the policy in force.
func (s *execState) current() Policy {
	if s == nil {
		return DenyAll() // an interpreter assembled without NewBareInterpreter fails closed
	}
	return *s.policy.Load()
}

// context returns the context bounding the run; context.Background when none is installed.
func (s *execState) context() context.Context {
	if s == nil {
		return context.Background()
	}
	if b := s.ctx.Load(); b != nil {
		return b.ctx
	}
	return context.Background()
}

// cancelled returns a *CancelError once the context of the run is done, else nil.
func (s *execState) cancelled() error {
	ctx := s.context()
	if ctx.Done() == nil || ctx.Err() == nil {
		return nil
	}
	return &CancelError{Cause: cancelCause(ctx)}
}

// interruptEvery is how many closure calls pass between context checks.
const interruptEvery = 256

type interrupter struct{ s *execState }

// Interrupted samples the run's context every interruptEvery calls.
func (it *interrupter) Interrupted() error {
	if it.s.calls.Add(1)%interruptEvery != 0 {
		return nil
	}
	return it.s.cancelled()
}

func (s *execState) interrupter() functions.Interrupter { return &interrupter{s: s} }

// SetPolicy installs p for this interpreter and every fork of it, and returns the policy it
// replaced so a caller can restore it.
func (i *Interpreter) SetPolicy(p Policy) (previous Policy) {
	if i.exec == nil {
		i.exec = newExecState()
	}
	previous = i.exec.current()
	i.exec.policy.Store(&p)
	return previous
}

// Policy returns the policy in force.
func (i *Interpreter) Policy() Policy { return i.exec.current() }

// RequireCapability returns a *CapabilityError when the policy in force denies c. Steps an
// Edition or plug-in registers call it before a side effect to take part in the policy:
//
//	if err := interp.RequireCapability(cli.CapEnv, "getenv", ""); err != nil { return failure(err) }
func (i *Interpreter) RequireCapability(c Capability, step, detail string) error {
	return i.exec.current().Require(c, step, detail)
}

// SetContext bounds the run by ctx for this interpreter and every fork, and returns the
// previous context. Cancellation is cooperative: every pipeline step checks it between
// nodes, the top-level loop between statements, and an OS command started by a step is
// killed. A single non-terminating evaluation of one element is not interruptible.
func (i *Interpreter) SetContext(ctx context.Context) (previous context.Context) {
	if i.exec == nil {
		i.exec = newExecState()
	}
	previous = i.exec.context()
	if ctx == nil {
		ctx = context.Background()
	}
	i.exec.ctx.Store(&ctxBox{ctx})
	return previous
}

// runContext is the context a step reads when its stream runs (a func, so an adapter built
// once and cached sees the context of the run it executes in).
func (i *Interpreter) runContext() context.Context { return i.exec.context() }

// CancelError is the failure of a pipeline stopped by its context: a deadline, a cancel, or
// the output limit. Cause is the context's cause (context.DeadlineExceeded, context.Canceled,
// ErrOutputLimit).
type CancelError struct{ Cause error }

// Error says why the run stopped.
func (e *CancelError) Error() string { return "the run was stopped: " + e.Cause.Error() }

// Unwrap exposes the cause to errors.Is (errors.Is(err, context.DeadlineExceeded)).
func (e *CancelError) Unwrap() error { return e.Cause }

// ErrOutputLimit is the cause of a run stopped because it printed more than RunOptions.MaxOutput bytes.
var ErrOutputLimit = errors.New("output limit exceeded")

func cancelCause(ctx context.Context) error {
	if c := context.Cause(ctx); c != nil {
		return c
	}
	return ctx.Err()
}

type nodeSeq = iter.Seq[dsltypes.ASTNode[functions.Either[error, any]]]
type pipelineStep = evaluator.PipelineStep[functions.Either[error, any]]

// gated wraps step so that it runs only when the policy in force grants c. Otherwise the
// stream is one error node naming the step and the capability; the wrapped step is never
// started. The check happens when the stream runs, not when the step is compiled.
func (i *Interpreter) gated(c Capability, name string, step pipelineStep) pipelineStep {
	return func(in nodeSeq) nodeSeq {
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			if err := i.exec.current().Require(c, name, ""); err != nil {
				yield(evaluator.InjectError[any](err, errorNodeFlag))
				return
			}
			for n := range step(in) {
				if !yield(n) {
					return
				}
			}
		}
	}
}

// GatedStep wraps an Edition's step handler so that every step it builds runs only when the
// policy in force grants c; otherwise the stream is one error node, `step 'NAME' requires
// the C capability (policy denies it)`, and the wrapped step is never started (Mandate 6: a
// denial is an error value, never an empty result). NAME is the first word of the step, so
// a step with an argument (`k8s_scale 3`) is reported as `k8s_scale`.
//
// Editions register steps through WithStep and those are, by default, the author's decision
// and always allowed. A step that WRITES to a remote system is the exception: wrap its
// handler, so the capability is checked when the stream runs (a policy tightened after the
// step was compiled still applies):
//
//	e.WithStep(cli.GatedStep(cli.CapClusterApply, applyStep))
func GatedStep(c Capability, handler StepHandler) StepHandler {
	return func(stepStr string, interp *Interpreter) (evaluator.PipelineStep[functions.Either[error, any]], bool) {
		step, ok := handler(stepStr, interp)
		if !ok {
			return nil, false
		}
		name, _, _ := strings.Cut(strings.TrimSpace(stepStr), " ")
		return interp.gated(c, name, step), true
	}
}

// cancellable wraps step so that its output stops, with one CancelError node, once the
// context of the run is done. With no context installed the step is returned unwrapped when
// its stream is built, so an uncancellable run pays one atomic load per step.
func (i *Interpreter) cancellable(step pipelineStep) pipelineStep {
	return func(in nodeSeq) nodeSeq {
		out := step(in)
		ctx := i.runContext()
		done := ctx.Done()
		if done == nil {
			return out
		}
		return func(yield func(dsltypes.ASTNode[functions.Either[error, any]]) bool) {
			select {
			case <-done: // already over: do not start the step at all
				yield(evaluator.InjectError[any](&CancelError{Cause: cancelCause(ctx)}, errorNodeFlag))
				return
			default:
			}
			for n := range out {
				if isCancelNode(n) {
					yield(n) // an earlier step already reported it; do not repeat
					return
				}
				select {
				case <-done:
					yield(evaluator.InjectError[any](&CancelError{Cause: cancelCause(ctx)}, errorNodeFlag))
					return
				default:
				}
				if !yield(n) {
					return
				}
			}
		}
	}
}

// errorNodeFlag is the type flag of the error nodes the wrappers yield: none. Looking a flag
// up by name registers it on first use and so is not safe from the goroutines `||>` starts;
// an error node is read from the Left channel and no step dispatches on its type.
const errorNodeFlag dsltypes.BitFlag = 0

func isCancelNode(n dsltypes.ASTNode[functions.Either[error, any]]) bool {
	if !n.Value.IsLeft() {
		return false
	}
	err, _ := n.Value.UnwrapLeft()
	var ce *CancelError
	return errors.As(err, &ce)
}
