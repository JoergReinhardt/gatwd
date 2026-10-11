package parser

import (
	"strings"
	"strconv"
	"math/big"
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

type Parser struct {
	typeRegistry map[string]dsltypes.BitFlag
	typeFactories map[string]func(string) any
	nextFlag     uint64
}

func NewParser() *Parser {
	return &Parser{
		typeRegistry: make(map[string]dsltypes.BitFlag),
		typeFactories: make(map[string]func(string) any),
		nextFlag:     1,
	}
}

func (p *Parser) Parse(script string) iter.Seq[dsltypes.ASTNode[any]] {
	return func(yield func(dsltypes.ASTNode[any]) bool) {
		lines := strings.Split(script, "\n")
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") { continue }

			if strings.HasPrefix(line, "type ") {
				p.parseType(line)
				continue
			}

			if strings.HasPrefix(line, "spawn ") {
				parts := strings.SplitN(line[6:], ":", 2)
				if len(parts) == 2 {
					typeName := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					
					flag, exists := p.typeRegistry[typeName]
					if exists {
						var parsedValue any = value
						// Try evaluating as infix expression first if it contains operators
						if strings.ContainsAny(value, "+-*/%&|<>= ") {
							if evaluated, err := EvalInfixString(value, nil); err == nil {
								parsedValue = evaluated
							} else if factory, hasFactory := p.typeFactories[typeName]; hasFactory {
								parsedValue = factory(value)
							}
						} else if factory, hasFactory := p.typeFactories[typeName]; hasFactory {
							parsedValue = factory(value)
						}
						node := dsltypes.NewASTNode[any](parsedValue, flag)
						if !yield(node) { return }
					}
				}
			}
		}
	}
}

func (p *Parser) registerOrGet(name string) dsltypes.BitFlag {
	if flag, exists := p.typeRegistry[name]; exists {
		return flag
	}
	newFlag := dsltypes.BitFlag(p.nextFlag)
	p.nextFlag <<= 1
	p.typeRegistry[name] = newFlag
	dsltypes.RegisterTypeFlag(newFlag, name)
	return newFlag
}

func (p *Parser) parseType(line string) {
	def := strings.TrimSpace(line[5:])
	parts := strings.SplitN(def, "=", 2)
	if len(parts) != 2 { return }

	typeName := strings.TrimSpace(parts[0])
	definition := strings.TrimSpace(parts[1])

	if strings.HasPrefix(definition, "{") && strings.HasSuffix(definition, "}") {
		// Record Type: { Field: Type, Field: Type }
		// Just register the base type for now (BitFlags can't inherently track struct shapes)
		p.registerOrGet(typeName)
	} else if (strings.HasPrefix(definition, "Set(") || strings.HasPrefix(definition, "List(") || strings.HasPrefix(definition, "Tree(") || strings.HasPrefix(definition, "BST(") || strings.HasPrefix(definition, "BTree(") || strings.HasPrefix(definition, "RBTree(")) && strings.HasSuffix(definition, ")") {
			p.registerOrGet(typeName)
		} else if strings.HasPrefix(definition, "(") && strings.HasSuffix(definition, ")") {
		// Tuple Type: (Type1, Type2)
		p.registerOrGet(typeName)
	} else if strings.Contains(definition, "|") {
		// Sum Type: Type1 | Type2
		subTypes := strings.Split(definition, "|")
		var composedFlag dsltypes.BitFlag = 0
		for _, sub := range subTypes {
			composedFlag = composedFlag.Concat(p.registerOrGet(strings.TrimSpace(sub)))
		}
		p.typeRegistry[typeName] = composedFlag
		dsltypes.RegisterTypeFlag(composedFlag, typeName)
	} else {
		// Base type
		if factory, ok := p.typeFactories[definition]; ok {
			p.typeFactories[typeName] = factory
		} else if definition == "BigInt" {
			p.typeFactories[typeName] = func(s string) any {
				n := new(big.Int)
				n.SetString(s, 10)
				return n
			}
		} else if definition == "BigFloat" {
			p.typeFactories[typeName] = func(s string) any {
				f, _, _ := new(big.Float).Parse(s, 10)
				return f
			}
		} else if definition == "Float" {
			p.typeFactories[typeName] = func(s string) any {
				f, _ := strconv.ParseFloat(s, 64)
				return f
			}
		} else if definition == "Int" {
			p.typeFactories[typeName] = func(s string) any {
				i, _ := strconv.Atoi(s)
				return i
			}
		} else if definition == "Uint" {
			p.typeFactories[typeName] = func(s string) any {
				u, _ := strconv.ParseUint(s, 10, 64)
				return u
			}
		} else if definition == "Complex" {
			p.typeFactories[typeName] = func(s string) any {
				c, _ := strconv.ParseComplex(s, 128)
				return c
			}
		}
		p.registerOrGet(typeName)
	}
}


// RegisterTypeFactory allows third-party Go code to register native constructors for DSL types.
// For example, RegisterTypeFactory("Time", func(s string) any { t, _ := time.Parse(time.RFC3339, s); return t })
func (p *Parser) RegisterTypeFactory(definition string, factory func(string) any) {
	p.typeFactories[definition] = factory
}
func (p *Parser) GetFlag(typeName string) dsltypes.BitFlag {
	return p.registerOrGet(typeName)
}
