package tt

// Node is the interface all AST nodes implement.
type Node interface {
	nodeType() string
}

// TemplateNode is the root of a parsed template.
type TemplateNode struct {
	Children []Node
}

func (n *TemplateNode) nodeType() string { return "Template" }

// TextNode represents raw template text.
type TextNode struct {
	Text string
}

func (n *TextNode) nodeType() string { return "Text" }

// GetNode represents [% expr %] or [% GET expr %].
type GetNode struct {
	Expr Expr
}

func (n *GetNode) nodeType() string { return "Get" }

// SetNode represents [% SET var = expr %] or [% var = expr %].
type SetNode struct {
	Pairs []SetPair
}

type SetPair struct {
	Variable Expr
	Value    Expr
}

func (n *SetNode) nodeType() string { return "Set" }

// DefaultNode represents [% DEFAULT var = expr %].
type DefaultNode struct {
	Pairs []SetPair
}

func (n *DefaultNode) nodeType() string { return "Default" }

// CallNode represents [% CALL expr %].
type CallNode struct {
	Expr Expr
}

func (n *CallNode) nodeType() string { return "Call" }

// IfNode represents IF/UNLESS/ELSIF/ELSE chains.
type IfNode struct {
	Branches []IfBranch
	Else     []Node
}

type IfBranch struct {
	Condition Expr
	Body      []Node
	Negate    bool // true for UNLESS
}

func (n *IfNode) nodeType() string { return "If" }

// ForeachNode represents [% FOREACH item IN list %] ... [% END %].
type ForeachNode struct {
	LoopVar string
	List    Expr
	Body    []Node
}

func (n *ForeachNode) nodeType() string { return "Foreach" }

// WhileNode represents [% WHILE expr %] ... [% END %].
type WhileNode struct {
	Condition Expr
	Body      []Node
}

func (n *WhileNode) nodeType() string { return "While" }

// SwitchNode represents [% SWITCH expr %] [% CASE ... %] ... [% END %].
type SwitchNode struct {
	Expr  Expr
	Cases []CaseBranch
}

type CaseBranch struct {
	Values []Expr // nil for default case
	Body   []Node
}

func (n *SwitchNode) nodeType() string { return "Switch" }

// BlockDefNode represents [% BLOCK name %] ... [% END %].
type BlockDefNode struct {
	Name string
	Body []Node
}

func (n *BlockDefNode) nodeType() string { return "BlockDef" }

// FilterBlockNode represents [% FILTER name %] ... [% END %].
type FilterBlockNode struct {
	Name string
	Args []Expr
	Body []Node
}

func (n *FilterBlockNode) nodeType() string { return "FilterBlock" }

// WrapperNode represents [% WRAPPER name %] ... [% END %].
type WrapperNode struct {
	Template Expr
	Params   []SetPair
	Body     []Node
}

func (n *WrapperNode) nodeType() string { return "Wrapper" }

// IncludeNode represents [% INCLUDE name key=val ... %].
type IncludeNode struct {
	Template Expr
	Params   []SetPair
	Localize bool // INCLUDE localizes vars; PROCESS does not
}

func (n *IncludeNode) nodeType() string { return "Include" }

// InsertNode represents [% INSERT filename %].
type InsertNode struct {
	Filename Expr
}

func (n *InsertNode) nodeType() string { return "Insert" }

// MacroNode represents [% MACRO name BLOCK %] ... [% END %]
// or [% MACRO name(args) BLOCK %] ... [% END %].
type MacroNode struct {
	Name   string
	Args   []string
	Body   []Node
}

func (n *MacroNode) nodeType() string { return "Macro" }

// TryNode represents [% TRY %] ... [% CATCH %] ... [% END %].
type TryNode struct {
	Body    []Node
	Catches []CatchBranch
	Final   []Node
}

type CatchBranch struct {
	Type string // error type or "" for default
	Body []Node
}

func (n *TryNode) nodeType() string { return "Try" }

// ThrowNode represents [% THROW type message %].
type ThrowNode struct {
	Type    Expr
	Message Expr
}

func (n *ThrowNode) nodeType() string { return "Throw" }

// FlowNode represents NEXT, LAST, RETURN, STOP, CLEAR.
type FlowNode struct {
	Action string // "NEXT", "LAST", "RETURN", "STOP", "CLEAR"
}

func (n *FlowNode) nodeType() string { return "Flow" }

// Expr is the interface for expression nodes.
type Expr interface {
	Node
	exprNode()
}

// IdentExpr represents a dotted variable reference: foo.bar.baz(args).
type IdentExpr struct {
	Segments []IdentSegment
}

type IdentSegment struct {
	Name    string
	Args    []Expr
	Dynamic bool // true when prefixed with $ — resolve Name as a variable to get the actual key
}

func (e *IdentExpr) nodeType() string { return "Ident" }
func (e *IdentExpr) exprNode()       {}

// StringExpr represents a string literal.
type StringExpr struct {
	Value        string
	Interpolated bool // double-quoted strings need interpolation
	Raw          string
}

func (e *StringExpr) nodeType() string { return "String" }
func (e *StringExpr) exprNode()       {}

// NumberExpr represents a numeric literal.
type NumberExpr struct {
	Value   string
	IsFloat bool
}

func (e *NumberExpr) nodeType() string { return "Number" }
func (e *NumberExpr) exprNode()       {}

// ArrayExpr represents [expr, expr, ...].
type ArrayExpr struct {
	Elements []Expr
}

func (e *ArrayExpr) nodeType() string { return "Array" }
func (e *ArrayExpr) exprNode()       {}

// HashExpr represents { key => val, ... }.
type HashExpr struct {
	Pairs []HashPair
}

type HashPair struct {
	Key   Expr
	Value Expr
}

func (e *HashExpr) nodeType() string { return "Hash" }
func (e *HashExpr) exprNode()       {}

// BinOpExpr represents binary operations: a OP b.
type BinOpExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

func (e *BinOpExpr) nodeType() string { return "BinOp" }
func (e *BinOpExpr) exprNode()       {}

// UnaryExpr represents unary operations: NOT x, !x.
type UnaryExpr struct {
	Op      string
	Operand Expr
}

func (e *UnaryExpr) nodeType() string { return "Unary" }
func (e *UnaryExpr) exprNode()       {}

// TernaryExpr represents cond ? then : else.
type TernaryExpr struct {
	Condition Expr
	Then      Expr
	Else      Expr
}

func (e *TernaryExpr) nodeType() string { return "Ternary" }
func (e *TernaryExpr) exprNode()       {}

// FilterExpr represents expr | filter(args).
type FilterExpr struct {
	Input  Expr
	Name   string
	Args   []Expr
}

func (e *FilterExpr) nodeType() string { return "Filter" }
func (e *FilterExpr) exprNode()       {}

// RangeExpr represents start..end.
type RangeExpr struct {
	Start Expr
	End   Expr
}

func (e *RangeExpr) nodeType() string { return "Range" }
func (e *RangeExpr) exprNode()       {}
