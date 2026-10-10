package parser

import (
	"strings"
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

// Parser handles translating the Guest DSL text into the AST Host Engine.
type Parser struct {
	// typeRegistry maps string names to dynamically allocated BitFlags
	typeRegistry map[string]dsltypes.BitFlag
	nextFlag     uint64
}

// NewParser creates a new ML-like DSL parser.
func NewParser() *Parser {
	return &Parser{
		typeRegistry: make(map[string]dsltypes.BitFlag),
		nextFlag:     1,
	}
}

// Parse runs a simple line-by-line interpreter over a DSL script.
//
// In our architecture (Phase 4), when the parser encounters a type definition
// (e.g., `type Config = Server | Client`), it updates the dynamic BitFlag 
// registry. When it encounters an evaluation command, it yields an ASTNode 
// directly into the Evaluator Trampoline.
func (p *Parser) Parse(script string) iter.Seq[dsltypes.ASTNode[string]] {
	return func(yield func(dsltypes.ASTNode[string]) bool) {
		lines := strings.Split(script, "\n")
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Simulated AST rule: Type definitions
			if strings.HasPrefix(line, "type ") {
				p.parseType(line)
				continue
			}

			// Simulated AST rule: Node Instantiation (e.g., "spawn Server: 10.0.0.1")
			if strings.HasPrefix(line, "spawn ") {
				parts := strings.Split(line[6:], ":")
				if len(parts) == 2 {
					typeName := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					
					flag, exists := p.typeRegistry[typeName]
					if exists {
						node := dsltypes.NewASTNode(value, flag)
						if !yield(node) {
							return
						}
					}
				}
			}
		}
	}
}

// parseType dynamically allocates BitFlags for Algebraic Data Types.
func (p *Parser) parseType(line string) {
	// strip "type "
	def := strings.TrimSpace(line[5:])
	parts := strings.Split(def, "=")
	if len(parts) != 2 {
		return
	}

	typeName := strings.TrimSpace(parts[0])
	definition := strings.TrimSpace(parts[1])

	// If it's a Sum type (e.g., Server | Client)
	if strings.Contains(definition, "|") {
		subTypes := strings.Split(definition, "|")
		var composedFlag dsltypes.BitFlag = 0

		for _, sub := range subTypes {
			sub = strings.TrimSpace(sub)
			if flag, exists := p.typeRegistry[sub]; exists {
				composedFlag = composedFlag.Concat(flag)
			} else {
				// Allocate new flag for unknown subtype
				newFlag := dsltypes.BitFlag(p.nextFlag)
				p.nextFlag <<= 1
				p.typeRegistry[sub] = newFlag
				composedFlag = composedFlag.Concat(newFlag)
			}
		}
		p.typeRegistry[typeName] = composedFlag
	} else {
		// Base type
		newFlag := dsltypes.BitFlag(p.nextFlag)
		p.nextFlag <<= 1
		p.typeRegistry[typeName] = newFlag
	}
}

// GetFlag retrieves the BitFlag for a given type name.
// This is useful for the Go host engine to register adapters manually.
func (p *Parser) GetFlag(typeName string) dsltypes.BitFlag {
	flag, exists := p.typeRegistry[typeName]
	if !exists {
		// If not found, allocate a new one (or return 0 depending on strictness)
		newFlag := dsltypes.BitFlag(p.nextFlag)
		p.nextFlag <<= 1
		p.typeRegistry[typeName] = newFlag
		return newFlag
	}
	return flag
}
