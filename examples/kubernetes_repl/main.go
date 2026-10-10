package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"iter"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/evaluator"
	"github.com/joergreinhardt/gatwd/interactive"
	"github.com/joergreinhardt/gatwd/parser"
)

func main() {
	fmt.Println("=== gatwd Interactive Kubernetes REPL ===")

	// 1. Initialize Kubernetes Dynamic Client using ~/.kube/config
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Failed to load kubeconfig: %v\n", err)
		os.Exit(1)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create dynamic client: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[+] Successfully connected to Kubernetes Cluster")

	// 2. Setup the DSL Parser and define our Types
	script := `
type KubeNamespace = Namespace
type KubePod = Pod
`
	p := parser.NewParser()
	p.Parse(script) // Populate the BitFlag registry

	// 3. Initialize Evaluator Engine
	engine := evaluator.NewEngine[string]()

	// 4. Register Native Kubernetes Adapter for Namespaces -> Pods
	// When evaluating a Namespace node, fetch all Pods in that namespace!
	namespaceFlag := p.GetFlag("KubeNamespace")
	podFlag := p.GetFlag("KubePod")
	
	podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	engine.Register(namespaceFlag, adapters.NativeKubernetesAdapter(client, podGVR, podFlag))
	fmt.Println("[+] Native Adapter Registered (KubeNamespace -> KubePod)")

	// 5. Setup the Interactive REPL (Phase 5)
	// When the engine hits a Pod, it won't have an adapter. It will pause and ask the user!
	engine.OnMissingAdapter = func(flag dsltypes.BitFlag) evaluator.Adapter[string] {
		fmt.Printf("\n[REPL PAUSE] Discovered nodes of Type %d, but no adapter exists to drill down further.\n", flag)
		fmt.Print("Enter a command to execute on these nodes (e.g. 'kubectl logs %s'): ")
		
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			fmt.Println("Skipping...")
			return nil
		}

		// The user creates a new generic Exec Adapter dynamically!
		// Output lines will be tagged as "Unknown" (BitFlag 0) for simplicity.
		return func(node dsltypes.ASTNode[string]) iter.Seq[dsltypes.ASTNode[string]] {
				return func(yield func(dsltypes.ASTNode[string]) bool) {
					anyAdapter := interactive.GenericExecAdapter(input, 0)
					anyNode := dsltypes.NewASTNode[any](node.Value, node.DSLType)
					for res := range anyAdapter(anyNode) {
						if s, ok := res.Value.(string); ok {
							yield(dsltypes.NewASTNode(s, res.DSLType))
						}
					}
				}
			}
	}

	// 6. Provide the Root Node (Target Namespace) to begin the crawl
	fmt.Print("\nEnter a Kubernetes Namespace to crawl (e.g., 'kube-system' or 'default'): ")
	reader := bufio.NewReader(os.Stdin)
	targetNs, _ := reader.ReadString('\n')
	targetNs = strings.TrimSpace(targetNs)

	fmt.Printf("\n--- Spawning Graph Crawler on Namespace: %s ---\n", targetNs)
	
	initialStream := func(yield func(dsltypes.ASTNode[string]) bool) {
		yield(dsltypes.NewASTNode(targetNs, namespaceFlag))
	}

	// 7. Evaluate the Graph! 
	// This will use the Native API to fetch Pods, and then pause to ask you what to do with the Pods.
	terminalStream := engine.Evaluate(initialStream)

	for node := range terminalStream {
		fmt.Printf("[Terminal Yield] -> %s\n", node.Value)
	}

	fmt.Println("--- Graph Traversal Complete ---")
}
