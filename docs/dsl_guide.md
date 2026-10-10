# Gophers All The Way Down (gatwd) - DSL Guide & Tutorial

Welcome to `gatwd`! This library provides an **Interactive, Functional Infrastructure Graph Explorer**. It is built on modern Go 1.23 Iterators and pure Category Theory principles, but you don't need a math degree to use it.

This guide will teach you the syntax of the DSL, how the interactive graph crawler works, and how to run it yourself.

---

## 1. What is the gatwd DSL?

The `gatwd` DSL is a tiny, ML-like scripting language designed to **declare infrastructure schemas on the fly** and **spawn graph traversal jobs**. 

It uses a mathematical concept called **Algebraic Data Types**. Instead of writing massive YAML files or Go structs, you define types dynamically by combining them.

### Simple Example: Defining Types
```text
// A Server can be either Linux OR Windows (A Sum Type)
type Server = LinuxHost | WindowsHost

// A Client is an Android or iOS device
type Client = Android | iOS

// An Endpoint can be either a Server OR a Client
type Endpoint = Server | Client
```
Because the engine uses a custom `BitFlag` registry, these types are parsed and registered dynamically at runtime. You do not need to recompile the Go host engine to add new types!

### Simple Example: Spawning Nodes
Once a type is defined, you can `spawn` a root node to kick off the graph traversal.

```text
// Spawns a root node representing a Linux host at the given IP
spawn LinuxHost: 192.168.0.1
```

---

## 2. How the Graph Crawler Works (Adapters)

When you `spawn` a node in the DSL, the engine tries to evaluate it. It does this by looking for an **Adapter** (a Type Class) registered to that specific type.

An Adapter is simply a Go function (or a CLI command wrapper) that takes the node (e.g., the IP address), performs an action (like running SSH), and returns a stream of *new* nodes (like a list of running Docker containers).

### Elaborate Example: Crawling Kubernetes
Imagine you want to crawl a Kubernetes cluster. 
1. You write a DSL script to define your infrastructure.
2. You register Adapters for those types.
3. The engine recursively crawls the graph!

**The DSL Script (`k8s.gatwd`):**
```text
type Cluster = KubernetesCluster
type Node = KubeNode
type Pod = KubePod

spawn KubernetesCluster: my-prod-cluster
```

**The Go Execution Code:**
```go
package main

import (
    "fmt"
    "github.com/joergreinhardt/gatwd/dsltypes"
    "github.com/joergreinhardt/gatwd/evaluator"
    "github.com/joergreinhardt/gatwd/parser"
    "github.com/joergreinhardt/gatwd/interactive"
)

func main() {
    // 1. Parse the script
    script := `
type Cluster = KubernetesCluster
type Node = KubeNode
type Pod = KubePod
spawn KubernetesCluster: my-prod-cluster`

    p := parser.NewParser()
    astStream := p.Parse(script)

    // 2. Initialize the Evaluator Engine
    engine := evaluator.NewEngine[string]()

    // 3. Register tool-agnostic CLI Adapters!
    // When the engine hits a KubernetesCluster, it runs `kubectl get nodes` 
    // and turns the output into `KubeNode`s.
    clusterFlag := p.GetFlag("KubernetesCluster")
    nodeFlag := p.GetFlag("KubeNode")
    engine.Register(clusterFlag, interactive.GenericExecAdapter("kubectl get nodes --context %s", nodeFlag))

    // When the engine hits a KubeNode, it gets the pods running on it.
    podFlag := p.GetFlag("KubePod")
    engine.Register(nodeFlag, interactive.GenericExecAdapter("kubectl get pods --field-selector spec.nodeName=%s", podFlag))

    // 4. Run the graph crawler!
    // Because this uses iter.Seq, it evaluates completely lazily and won't blow up your RAM.
    terminalNodes := engine.Evaluate(astStream)

    for node := range terminalNodes {
        fmt.Println("Found Pod:", node.Value)
    }
}
```

---

## 3. The Killer Feature: The Interactive REPL

What happens if the engine encounters a node it doesn't have an Adapter for? In a normal script, the program crashes. In `gatwd`, the engine **pauses**.

Because the evaluator is built on lazy sequences (the Trampoline), it can safely suspend execution, prompt you (the user) for a CLI command, compile an Adapter on the fly, and resume exactly where it left off.

**How to enable it:**
```go
engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[string] {
    fmt.Printf("Encountered unknown Node Type %d. Enter a CLI command to explore it (e.g., 'ssh %%s uptime'): ", flag)
    
    // Read from standard input
    reader := bufio.NewReader(os.Stdin)
    cmd, _ := reader.ReadString('\n')
    cmd = strings.TrimSpace(cmd)
    
    // Generate a generic adapter on the fly!
    return interactive.GenericExecAdapter(cmd, dsltypes.BitFlag(0))
}
```

---

## 4. Advanced Concepts under the Hood

If you are a library developer, you have access to cutting-edge Category Theory mechanics to manipulate the graph safely:

* **Zippers (`comonads` package):** Need to navigate up and down an immutable parsed AST? `TreeZipper` provides O(1) Cursor-like navigation, allowing you to write contextual DSL rules (e.g., "If this node has a LoadBalancer parent, do X").
* **Optics (`optics` package):** Need to modify a deeply nested field in an immutable configuration tree? Use `Lens` to functionally set the value and return a brand new tree without mutating the original.
* **Parallel Execution (`functions` package):** Want to crawl 10,000 servers at once? Simply replace `engine.Evaluate` with `engine.ParEvaluate(stream, 100)`. It uses the `ParBind` Monad to safely parallelize the graph crawler across 100 goroutines!
