package parser

import (
	"iter"
	"math/big"
	"strconv"
	"strings"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

// Parser compiles Gatwd type declarations into runtime metadata and turns a script
// into a lazy stream of typed AST nodes. It maintains three registries: typeRegistry
// maps each declared type name to a unique dsltypes.BitFlag used as its runtime tag,
// typeFactories maps a type name to a native Go constructor that parses a literal
// string into a concrete value (e.g. "Int" to strconv.Atoi), and sumTypes maps each
// declared sum (algebraic data) type to its branch definitions. nextFlag holds the
// next power-of-two bit to hand out, so every type gets a distinct single-bit flag
// that can be unioned (Concat) to describe a sum type as the OR of its branches.
type Parser struct {
	typeRegistry  map[string]dsltypes.BitFlag
	typeFactories map[string]func(string) any
	sumTypes      map[string]*dsltypes.SumTypeDefinition
	nextFlag      uint64
}

// NewParser returns a Parser with empty registries and the first BitFlag bit (1)
// ready to be allocated. Each subsequent registered type shifts nextFlag left by one,
// so the Parser can tag up to 64 distinct base types with single-bit flags.
func NewParser() *Parser {
	return &Parser{
		typeRegistry:  make(map[string]dsltypes.BitFlag),
		typeFactories: make(map[string]func(string) any),
		sumTypes:      make(map[string]*dsltypes.SumTypeDefinition),
		nextFlag:      1,
	}
}

// Parse reads a Gatwd script and returns a lazy iter.Seq that yields one
// dsltypes.ASTNode for every `spawn` statement in source order. It processes the
// script line by line: blank lines and `--` comments are skipped, `type ...` lines
// register a type via parseType (producing no node), and `spawn <Type>: <value>`
// lines look up the named type's BitFlag and box the value into an ASTNode tagged
// with that flag. The value is resolved in priority order: if it contains operator
// characters it is first evaluated as an infix expression, otherwise a registered
// type factory parses it, and a bare double-quoted literal is unquoted; finally, if
// the type is a sum type and the value names one of its branches, MakeVariant boxes
// it as a dsltypes.Variant. The returned sequence is lazy, so parsing of later lines
// only happens as the caller pulls nodes, and iteration stops early if yield returns
// false. Parse itself yields no node for a spawn of a type that has not been declared
// (there is no flag to tag it with); the interpreter in pkg/cli registers a spawned
// type on demand before the line reaches this point, so scripts never see that.
func (p *Parser) Parse(script string) iter.Seq[dsltypes.ASTNode[any]] {
	return func(yield func(dsltypes.ASTNode[any]) bool) {
		lines := strings.Split(script, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") {
				continue
			}

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
						} else if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
							parsedValue = value[1 : len(value)-1]
						}
						if variant, ok := p.MakeVariant(typeName, value); ok {
							parsedValue = variant
						}
						node := dsltypes.NewASTNode[any](parsedValue, flag)
						if !yield(node) {
							return
						}
					}
				}
			}
		}
	}
}

// MakeVariant attempts to interpret valStr as a constructor of the named sum type.
// It returns the boxed Variant and true when typeName is a declared sum type and
// valStr names one of its branches (optionally with a "(payload)" suffix, e.g.
// "Just(1 + 2)"); otherwise it returns the zero Variant and false. Both the
// top-level Parse path and the interpreter's spawn path use this so that a value
// like `spawn OS: Linux` boxes a dsltypes.Variant{Tag:"Linux"} that the match
// evaluator can discriminate on.
func (p *Parser) MakeVariant(typeName, valStr string) (dsltypes.Variant, bool) {
	sumDef, isSum := p.sumTypes[typeName]
	if !isSum {
		return dsltypes.Variant{}, false
	}
	branchTag := valStr
	var payload any = nil
	if parenIdx := strings.Index(valStr, "("); parenIdx != -1 && strings.HasSuffix(valStr, ")") {
		branchTag = strings.TrimSpace(valStr[:parenIdx])
		rawInner := strings.TrimSpace(valStr[parenIdx+1 : len(valStr)-1])
		if evalInfix, err := EvalInfixString(rawInner, nil); err == nil {
			payload = evalInfix
		} else {
			payload = rawInner
		}
	}
	if branchFlag, hasBranch := sumDef.Branches[branchTag]; hasBranch {
		return dsltypes.NewVariant(branchTag, branchFlag, payload), true
	}
	return dsltypes.Variant{}, false
}

// SumTypeOf returns the name of the sum type that declares the given constructor
// branch and all of that type's branch names, or ok=false if name is not a
// registered constructor. The match exhaustiveness check in `gatwd --check` uses
// this to determine which branches of a closed sum type a `match` block leaves
// uncovered.
func (p *Parser) SumTypeOf(branch string) (name string, branches []string, ok bool) {
	for typeName, sumDef := range p.sumTypes {
		if _, has := sumDef.Branches[branch]; has {
			all := make([]string, 0, len(sumDef.Branches))
			for b := range sumDef.Branches {
				all = append(all, b)
			}
			return typeName, all, true
		}
	}
	return "", nil, false
}

// IsConstructor reports whether name is a registered constructor (branch) of any
// declared sum type. The match evaluator uses this to distinguish a pattern that
// names an ADT constructor (which must discriminate on the value's variant tag)
// from a pattern that introduces a fresh variable binding (which matches anything).
func (p *Parser) IsConstructor(name string) bool {
	for _, sumDef := range p.sumTypes {
		if _, ok := sumDef.Branches[name]; ok {
			return true
		}
	}
	return false
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
	if len(parts) != 2 {
		return
	}

	typeName := strings.TrimSpace(parts[0])
	definition := strings.TrimSpace(parts[1])

	if strings.HasPrefix(definition, "{") && strings.HasSuffix(definition, "}") {
		// Record Type: { Field: Type, Field: Type }
		// Just register the base type for now (BitFlags can't inherently track struct shapes)
		p.registerOrGet(typeName)
	} else if (strings.HasPrefix(definition, "Set(") || strings.HasPrefix(definition, "List(") || strings.HasPrefix(definition, "Grid(") || strings.HasPrefix(definition, "Graph(") || strings.HasPrefix(definition, "Tree(") || strings.HasPrefix(definition, "BST(") || strings.HasPrefix(definition, "BTree(") || strings.HasPrefix(definition, "RBTree(")) && strings.HasSuffix(definition, ")") {
		p.registerOrGet(typeName)
	} else if strings.HasPrefix(definition, "(") && strings.HasSuffix(definition, ")") {
		// Tuple Type: (Type1, Type2)
		p.registerOrGet(typeName)
	} else if strings.Contains(definition, "|") {
		// Sum Type: Type1 | Type2
		subTypes := strings.Split(definition, "|")
		var composedFlag dsltypes.BitFlag = 0
		sumDef := dsltypes.NewSumTypeDefinition(typeName, 0)
		for _, sub := range subTypes {
			branch := strings.TrimSpace(sub)
			branchFlag := p.registerOrGet(branch)
			composedFlag = composedFlag.Concat(branchFlag)
			sumDef.AddBranch(branch, branchFlag)
		}
		sumDef.Flag = composedFlag
		p.typeRegistry[typeName] = composedFlag
		p.sumTypes[typeName] = sumDef
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

// GetFlag returns the dsltypes.BitFlag tag for typeName, allocating and registering a
// fresh single-bit flag the first time the name is seen. Because it registers on
// demand, callers can obtain a stable runtime tag for a type even before (or without)
// a full `type` declaration.
func (p *Parser) GetFlag(typeName string) dsltypes.BitFlag {
	return p.registerOrGet(typeName)
}

// GetFactory returns the native constructor registered for typeName and reports
// whether one exists. The factory converts a raw literal string into a concrete Go
// value (for example parsing "42" into an int); ok is false for types that have no
// factory, such as sum types or bare base types that are stored as their string form.
func (p *Parser) GetFactory(typeName string) (func(string) any, bool) {
	f, ok := p.typeFactories[typeName]
	return f, ok
}
