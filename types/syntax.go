package types

import "text/scanner"

///// SYNTAX DEFINITION /////
type TokType BitFlag

func (t TokType) Flag() BitFlag  { return BitFlag(t) }
func (t TokType) Syntax() string { return syntax[t] }

//go:generate stringer -type=TokType
const (
	tok_none  TokType = 1
	tok_blank TokType = 1 << iota
	tok_underscore
	tok_asterisk
	tok_dot
	tok_comma
	tok_colon
	tok_semicolon
	tok_minus
	tok_plus
	tok_or
	tok_xor
	tok_and
	tok_equal
	tok_lesser
	tok_greater
	tok_leftPar
	tok_rightPar
	tok_leftBra
	tok_rightBra
	tok_leftCur
	tok_rightCur
	tok_slash
	tok_not
	tok_dec
	tok_inc
	tok_doubEqual
	tok_rightArrow
	tok_leftArrow
	tok_fatLArrow
	tok_fatRArrow
	tok_doubCol
	tok_sing_quote
	tok_doub_quote
	tok_bckSla
	tok_lambda
	tok_number
	tok_letter
	tok_capital
	tok_genType
	tok_headWord
	tok_tailWord
	tok_inWord
	tok_conWord
	tok_letWord
	tok_whereWord
	tok_otherwiseWord
	tok_ifWord
	tok_thenWord
	tok_elseWord
	tok_caseWord
	tok_ofWord
	tok_dataWord
	tok_typeWord
	tok_typeIdent
	tok_funcIdent
)

var syntax = map[TokType]string{
	tok_none:          "",
	tok_blank:         " ",
	tok_underscore:    "_",
	tok_asterisk:      "*",
	tok_dot:           ".",
	tok_comma:         ",",
	tok_colon:         ":",
	tok_semicolon:     ";",
	tok_minus:         "-",
	tok_plus:          "+",
	tok_or:            "|",
	tok_xor:           "^",
	tok_and:           "&",
	tok_equal:         "=",
	tok_lesser:        "<",
	tok_greater:       ">",
	tok_leftPar:       "(",
	tok_rightPar:      ")",
	tok_leftBra:       "[",
	tok_rightBra:      "]",
	tok_leftCur:       "{",
	tok_rightCur:      "}",
	tok_slash:         "/",
	tok_not:           "&^",
	tok_dec:           "--",
	tok_inc:           "++",
	tok_doubEqual:     "==",
	tok_rightArrow:    "->",
	tok_leftArrow:     "<-",
	tok_fatLArrow:     "<=",
	tok_fatRArrow:     "=>",
	tok_doubCol:       "::",
	tok_sing_quote:    `'`,
	tok_doub_quote:    `"`,
	tok_bckSla:        `\`,
	tok_lambda:        `\x`,
	tok_number:        "[0-9]",
	tok_letter:        "[a-z]",
	tok_capital:       "[A-Z]",
	tok_genType:       "[[a-w]|y|z]",
	tok_headWord:      "x",
	tok_tailWord:      "xs",
	tok_inWord:        "in",
	tok_conWord:       "con",
	tok_letWord:       "let",
	tok_whereWord:     "where",
	tok_otherwiseWord: "otherwise",
	tok_ifWord:        "if",
	tok_thenWord:      "then",
	tok_elseWord:      "else",
	tok_caseWord:      "case",
	tok_ofWord:        "of",
	tok_dataWord:      "data",
	tok_typeWord:      "type",
	tok_typeIdent:     "[A-z][a-z]*",
	tok_funcIdent:     "([a-w|y|z][a-z])|(x[a-r|t-z])",
}

//// item type according to text, scanner tokenizer.
type item struct {
	typ  rune
	text string
	pos  scanner.Position
}

func newItem(typ rune, text string, pos scanner.Position) item {
	return item{typ, text, pos}
}
func (t item) Text() string           { return t.text }
func (t item) ItemType() rune         { return t.typ }
func (t item) ItemTypeString() string { return scanner.TokenString(t.typ) }

//////// IDENTITY & TYPE-REGISTER TYPES /////////
//go:generate stringer -type NodeT
type NodeT BitFlag

func (n NodeT) Flag() BitFlag { return BitFlag(n) }

const (
	NodeRoot  NodeT = 0
	NodeChain NodeT = 1 + iota
	NodeBranch
	NodeNest
	NodeLeave
)

// all user defined types get registered, indexed and mapped by name
type typeIdx []typeDef
type typeReg map[string]typeDef

var typeIndex typeIdx
var typeRegister typeReg

func initTypeDef() {
	typeIndex = []typeDef{}
	typeRegister = map[string]typeDef{}
}

type typeDef struct {
	Id    int          // id == own index position in typeIdx
	Princ BitFlag      // <-- principle type
	Name  string       // <-- name of this type
	Deri  []int        // <-- id's of derived types
	Fnc   []Functional // <-- constructors (type&data)
	next  Nodular
}

func (td typeDef) Eval() Data        { return td }
func (td typeDef) String() string    { return td.Name }
func (td typeDef) Flag() BitFlag     { return Node.Flag() }
func (td typeDef) NodeType() BitFlag { return NodeRoot.Flag() }
func (td typeDef) Root() Nodular     { return nil }
func (td typeDef) Empty() bool {
	if td.next != nil {
		return false
	}
	return true
}
func (td *typeDef) Next() Nodular { return (*td).next }
func (td *typeDef) Derive(princ BitFlag, name string, flags ...BitFlag) *typeDef {
	var dtd = conTypeDef(princ, name, flags...)
	(*td).Deri = append(td.Deri, dtd.Id)
	return dtd
}
func conTypeDef(princ BitFlag, name string, flags ...BitFlag) *typeDef {
	var id = len(typeIndex)
	var fid = BitFlag(id)
	var highfid = fhigh(fid)
	var pt = fconc(highfid, princ)
	var td = typeDef{
		Id:    id,
		Princ: pt,
		Name:  name,
		Deri:  []int{},
		Fnc:   []Functional{},
		next:  nil,
	}
	if len(flags) > 0 {
		node := *conChainNode(&td, flags...)
		td.next = &node
	}
	return &td
}

type leaveSigNode struct {
	Tok  Typed
	Text string
	root *typeDef
}

func (l leaveSigNode) Eval() Data        { return l }
func (l leaveSigNode) Flag() BitFlag     { return Node.Flag() }
func (l leaveSigNode) NodeType() BitFlag { return NodeLeave.Flag() }
func (l *leaveSigNode) Root() Nodular    { return (*l).root }
func (l leaveSigNode) Empty() bool {
	if l.Text != "" {
		return false
	}
	return true
}
func (l leaveSigNode) String() string {
	if f, ok := l.Tok.(Type); ok {
		return f.String()
	}
	return l.Tok.(TokType).String()

}
func conLeaveNode(root *typeDef, tok BitFlag) *leaveSigNode {
	var text string
	return &leaveSigNode{
		Tok:  tok,
		Text: text,
		root: root,
	}
}

type chainSigNode struct {
	*leaveSigNode
	next Nodular
}

func (c chainSigNode) Eval() Data        { return c }
func (c chainSigNode) NodeType() BitFlag { return NodeChain.Flag() }
func (c *chainSigNode) Next() Nodular    { return (*c).next }
func (c chainSigNode) Empty() bool {
	if c.next != nil {
		return false
	}
	return true
}
func conChainNode(root *typeDef, d ...BitFlag) *chainSigNode {
	r := root
	var csn chainSigNode
	switch len(d) {
	case 0:
		return nil
	case 1:
		csn = chainSigNode{conLeaveNode(r, d[0]), nil}
	default:
		var next = *conChainNode(r, d[1:]...)
		csn = chainSigNode{conLeaveNode(r, d[0]), &next}
	}
	return &csn
}

type branchSigNode struct {
	*leaveSigNode
	Left  Nodular
	Right Nodular
}

func (c branchSigNode) Eval() Data        { return c }
func (c branchSigNode) NodeType() BitFlag { return NodeBranch.Flag() }
func (c branchSigNode) Empty() bool {
	if (c.Left != nil) && (c.Right != nil) {
		if fmatch(c.Left.Flag(), Nil.Flag()) && fmatch(c.Right.Flag(), Nil.Flag()) {
			return false
		}
	}
	return true
}

type nestSigNode struct {
	*leaveSigNode
	member []Nodular
}

func (s nestSigNode) Eval() Data         { return s }
func (s nestSigNode) NodeType() BitFlag  { return NodeNest.Flag() }
func (s *nestSigNode) Member() []Nodular { return (*s).member }
func (l nestSigNode) Empty() bool {
	if len(l.member) > 0 {
		for _, m := range l.member {
			if !m.Empty() {
				return false
			}
		}
	}
	return true
}
