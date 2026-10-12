package dsltypes

import "fmt"

// Variant represents a value of a tagged union / sum type.
// It stores the specific branch tag, the branch's BitFlag, and the payload value.
type Variant struct {
	Tag   string  // Branch name, e.g. "Just", "Nothing", "Active", "Paused", "Error"
	Flag  BitFlag // The branch or sum BitFlag
	Value any     // The associated payload (may be nil for unit variants)
}

// NewVariant constructs a tagged variant instance.
func NewVariant(tag string, flag BitFlag, val any) Variant {
	return Variant{
		Tag:   tag,
		Flag:  flag,
		Value: val,
	}
}

// Is checks if the variant matches a given branch tag.
func (v Variant) Is(tag string) bool {
	return v.Tag == tag
}

// MatchType checks if the variant's flag matches the expected BitFlag.
func (v Variant) MatchType(expected BitFlag) bool {
	return v.Flag.Match(expected)
}

// String implements fmt.Stringer for clean output representation.
func (v Variant) String() string {
	if v.Value == nil {
		return v.Tag
	}
	return fmt.Sprintf("%s(%v)", v.Tag, v.Value)
}

// VariantElements exposes tag, bitflag (as uint64), and payload for structural comparison.
func (v Variant) VariantElements() (string, uint64, any) {
	return v.Tag, uint64(v.Flag), v.Value
}

// Equal checks equality of two variants by tag, flag, and payload.
func (v Variant) Equal(other Variant) bool {
	if v.Tag != other.Tag || v.Flag != other.Flag {
		return false
	}
	if v.Value == nil && other.Value == nil {
		return true
	}
	if v.Value == nil || other.Value == nil {
		return false
	}
	return fmt.Sprintf("%v", v.Value) == fmt.Sprintf("%v", other.Value)
}

// SumTypeDefinition captures the declared branches of an algebraic sum type.
type SumTypeDefinition struct {
	Name     string
	Flag     BitFlag
	Branches map[string]BitFlag // Branch name -> bit flag
}

// NewSumTypeDefinition creates a new sum type definition.
func NewSumTypeDefinition(name string, flag BitFlag) *SumTypeDefinition {
	return &SumTypeDefinition{
		Name:     name,
		Flag:     flag,
		Branches: make(map[string]BitFlag),
	}
}

// AddBranch registers a variant branch.
func (s *SumTypeDefinition) AddBranch(branchName string, flag BitFlag) {
	s.Branches[branchName] = flag
}

// HasBranch checks if a branch is part of this sum type.
func (s *SumTypeDefinition) HasBranch(branchName string) bool {
	_, ok := s.Branches[branchName]
	return ok
}
