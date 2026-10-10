package parser

import (
	"strings"
	"iter"

	"github.com/joergreinhardt/gatwd/dsltypes"
)

type Parser struct {
	typeRegistry map[string]dsltypes.BitFlag
	nextFlag     uint64
}

func NewParser() *Parser {
	return &Parser{
		typeRegistry: make(map[string]dsltypes.BitFlag),
		nextFlag:     1,
	}
}

func (p *Parser) Parse(script string) iter.Seq[dsltypes.ASTNode[string]] {
	return func(yield func(dsltypes.ASTNode[string]) bool) {
		lines := strings.Split(script, "\n")
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") { continue }

			if strings.HasPrefix(line, "type ") {
				p.parseType(line)
				continue
			}

			if strings.HasPrefix(line, "spawn ") {
				parts := strings.Split(line[6:], ":")
				if len(parts) == 2 {
					typeName := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					
					flag, exists := p.typeRegistry[typeName]
					if exists {
						node := dsltypes.NewASTNode(value, flag)
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
	} else if (strings.HasPrefix(definition, "Set(") || strings.HasPrefix(definition, "List(") || strings.HasPrefix(definition, "Tree(")) && strings.HasSuffix(definition, ")") {
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
	} else {
		// Base type
		p.registerOrGet(typeName)
	}
}

func (p *Parser) GetFlag(typeName string) dsltypes.BitFlag {
	return p.registerOrGet(typeName)
}
