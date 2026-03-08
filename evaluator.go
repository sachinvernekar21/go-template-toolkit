package tt

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
)

// Evaluator walks the AST and produces output.
type Evaluator struct {
	stash   *Stash
	output  io.Writer
	filters map[string]FilterFunc
	blocks  map[string]*BlockDefNode
	macros  map[string]*MacroNode
	loader  TemplateLoader
}

// TemplateLoader is the interface for loading included templates.
type TemplateLoader interface {
	Load(name string) (string, error)
}

func NewEvaluator(stash *Stash, output io.Writer, loader TemplateLoader) *Evaluator {
	filters := make(map[string]FilterFunc)
	for k, v := range defaultFilters {
		filters[k] = v
	}
	return &Evaluator{
		stash:   stash,
		output:  output,
		filters: filters,
		blocks:  make(map[string]*BlockDefNode),
		macros:  make(map[string]*MacroNode),
		loader:  loader,
	}
}

func (e *Evaluator) AddFilter(name string, fn FilterFunc) {
	e.filters[name] = fn
}

func (e *Evaluator) Execute(tmpl *TemplateNode) error {
	// First pass: collect block definitions
	for _, child := range tmpl.Children {
		if block, ok := child.(*BlockDefNode); ok {
			e.blocks[block.Name] = block
		}
		if macro, ok := child.(*MacroNode); ok {
			e.macros[macro.Name] = macro
		}
	}
	return e.execNodes(tmpl.Children)
}

func (e *Evaluator) execNodes(nodes []Node) error {
	for _, node := range nodes {
		if err := e.execNode(node); err != nil {
			return err
		}
	}
	return nil
}

func (e *Evaluator) execNode(node Node) error {
	switch n := node.(type) {
	case *TextNode:
		_, err := io.WriteString(e.output, n.Text)
		return err

	case *GetNode:
		val, err := e.evalExpr(n.Expr)
		if err != nil {
			return err
		}
		_, err = io.WriteString(e.output, toString(val))
		return err

	case *SetNode:
		return e.execSet(n.Pairs)

	case *DefaultNode:
		return e.execDefault(n.Pairs)

	case *CallNode:
		_, err := e.evalExpr(n.Expr)
		return err

	case *IfNode:
		return e.execIf(n)

	case *ForeachNode:
		return e.execForeach(n)

	case *WhileNode:
		return e.execWhile(n)

	case *SwitchNode:
		return e.execSwitch(n)

	case *BlockDefNode:
		e.blocks[n.Name] = n
		return nil

	case *FilterBlockNode:
		return e.execFilterBlock(n)

	case *WrapperNode:
		return e.execWrapper(n)

	case *IncludeNode:
		return e.execInclude(n)

	case *InsertNode:
		return e.execInsert(n)

	case *MacroNode:
		e.macros[n.Name] = n
		return nil

	case *TryNode:
		return e.execTry(n)

	case *ThrowNode:
		return e.execThrow(n)

	case *FlowNode:
		return &FlowSignal{Action: n.Action}

	default:
		return newEvalError(fmt.Sprintf("unknown node type: %T", node))
	}
}

func (e *Evaluator) execSet(pairs []SetPair) error {
	for _, pair := range pairs {
		val, err := e.evalExpr(pair.Value)
		if err != nil {
			return err
		}
		key := e.identKey(pair.Variable)
		if key != "" {
			e.stash.Set(key, val)
		}
	}
	return nil
}

func (e *Evaluator) execDefault(pairs []SetPair) error {
	for _, pair := range pairs {
		key := e.identKey(pair.Variable)
		if key == "" {
			continue
		}
		existing, ok := e.stash.Get(key)
		if ok && isTruthy(existing) {
			continue
		}
		val, err := e.evalExpr(pair.Value)
		if err != nil {
			return err
		}
		e.stash.Set(key, val)
	}
	return nil
}

func (e *Evaluator) identKey(expr Expr) string {
	if ident, ok := expr.(*IdentExpr); ok && len(ident.Segments) == 1 {
		return ident.Segments[0].Name
	}
	return ""
}

func (e *Evaluator) execIf(n *IfNode) error {
	for _, branch := range n.Branches {
		cond, err := e.evalExpr(branch.Condition)
		if err != nil {
			return err
		}
		truth := isTruthy(cond)
		if branch.Negate {
			truth = !truth
		}
		if truth {
			return e.execNodes(branch.Body)
		}
	}
	if n.Else != nil {
		return e.execNodes(n.Else)
	}
	return nil
}

// LoopInfo holds iteration metadata exposed as the `loop` variable inside FOREACH.
// Using a struct avoids collisions with hash virtual methods (e.g. "size").
type LoopInfo struct {
	Index int
	Count int
	Size  int
	Max   int
	First bool
	Last  bool
	Prev  string
	Next  string
}

func (e *Evaluator) execForeach(n *ForeachNode) error {
	listVal, err := e.evalExpr(n.List)
	if err != nil {
		return err
	}

	list := toSlice(listVal)

	oldStash := e.stash
	e.stash = oldStash.Clone()
	defer func() { e.stash = oldStash }()

	info := &LoopInfo{
		Size: len(list),
		Max:  len(list) - 1,
	}

	for i, item := range list {
		info.Index = i
		info.Count = i + 1
		info.First = i == 0
		info.Last = i == len(list) - 1
		info.Prev = ""
		info.Next = ""
		if i > 0 {
			info.Prev = toString(list[i-1])
		}
		if i < len(list)-1 {
			info.Next = toString(list[i+1])
		}

		e.stash.Set("loop", info)
		if n.LoopVar != "" {
			e.stash.Set(n.LoopVar, item)
		}

		err := e.execNodes(n.Body)
		if err != nil {
			if fs, ok := err.(*FlowSignal); ok {
				switch fs.Action {
				case "NEXT":
					continue
				case "LAST":
					return nil
				default:
					return err
				}
			}
			return err
		}
	}
	return nil
}

func (e *Evaluator) execWhile(n *WhileNode) error {
	const maxIter = 10000
	for i := 0; i < maxIter; i++ {
		cond, err := e.evalExpr(n.Condition)
		if err != nil {
			return err
		}
		if !isTruthy(cond) {
			break
		}
		err = e.execNodes(n.Body)
		if err != nil {
			if fs, ok := err.(*FlowSignal); ok {
				switch fs.Action {
				case "NEXT":
					continue
				case "LAST":
					return nil
				default:
					return err
				}
			}
			return err
		}
	}
	return nil
}

func (e *Evaluator) execSwitch(n *SwitchNode) error {
	val, err := e.evalExpr(n.Expr)
	if err != nil {
		return err
	}
	valStr := toString(val)

	for _, c := range n.Cases {
		if len(c.Values) == 0 {
			return e.execNodes(c.Body)
		}
		for _, cv := range c.Values {
			caseVal, err := e.evalExpr(cv)
			if err != nil {
				return err
			}
			if toString(caseVal) == valStr {
				return e.execNodes(c.Body)
			}
		}
	}
	return nil
}

func (e *Evaluator) execFilterBlock(n *FilterBlockNode) error {
	var buf strings.Builder
	oldOutput := e.output
	e.output = &buf
	err := e.execNodes(n.Body)
	e.output = oldOutput
	if err != nil {
		return err
	}

	result := buf.String()
	args, err := e.evalArgExprs(n.Args)
	if err != nil {
		return err
	}

	fn, ok := e.filters[n.Name]
	if !ok {
		return newEvalError(fmt.Sprintf("unknown filter: %s", n.Name))
	}
	result = fn(result, args)
	_, err = io.WriteString(e.output, result)
	return err
}

func (e *Evaluator) execWrapper(n *WrapperNode) error {
	var bodyBuf strings.Builder
	oldOutput := e.output
	e.output = &bodyBuf
	err := e.execNodes(n.Body)
	e.output = oldOutput
	if err != nil {
		return err
	}

	tmplName := e.resolveTemplateName(n.Template)

	if e.loader == nil {
		return newFileError("no template loader configured")
	}

	source, err := e.loader.Load(tmplName)
	if err != nil {
		return err
	}

	parsed, err := Parse(source, "", "")
	if err != nil {
		return err
	}

	oldStash := e.stash
	e.stash = oldStash.Clone()
	defer func() { e.stash = oldStash }()

	e.stash.Set("content", bodyBuf.String())
	for _, p := range n.Params {
		val, err := e.evalExpr(p.Value)
		if err != nil {
			return err
		}
		key := e.identKey(p.Variable)
		if key != "" {
			e.stash.Set(key, val)
		}
	}

	return e.Execute(parsed)
}

func (e *Evaluator) execInclude(n *IncludeNode) error {
	tmplName := e.resolveTemplateName(n.Template)

	// Check if it's a named block
	if block, ok := e.blocks[tmplName]; ok {
		if n.Localize {
			oldStash := e.stash
			e.stash = oldStash.Clone()
			defer func() { e.stash = oldStash }()
		}
		for _, p := range n.Params {
			val, err := e.evalExpr(p.Value)
			if err != nil {
				return err
			}
			key := e.identKey(p.Variable)
			if key != "" {
				e.stash.Set(key, val)
			}
		}
		return e.execNodes(block.Body)
	}

	if e.loader == nil {
		return newFileError(fmt.Sprintf("no template loader configured, cannot load %q", tmplName))
	}

	source, err := e.loader.Load(tmplName)
	if err != nil {
		return err
	}

	parsed, err := Parse(source, "", "")
	if err != nil {
		return err
	}

	if n.Localize {
		oldStash := e.stash
		e.stash = oldStash.Clone()
		defer func() { e.stash = oldStash }()
	}

	for _, p := range n.Params {
		val, err := e.evalExpr(p.Value)
		if err != nil {
			return err
		}
		key := e.identKey(p.Variable)
		if key != "" {
			e.stash.Set(key, val)
		}
	}

	return e.Execute(parsed)
}

func (e *Evaluator) execInsert(n *InsertNode) error {
	filename := e.resolveTemplateName(n.Filename)

	if e.loader == nil {
		return newFileError(fmt.Sprintf("no template loader configured, cannot insert %q", filename))
	}

	content, err := e.loader.Load(filename)
	if err != nil {
		return err
	}

	_, err = io.WriteString(e.output, content)
	return err
}

func (e *Evaluator) execTry(n *TryNode) error {
	err := e.execNodes(n.Body)
	if err != nil {
		if _, ok := err.(*FlowSignal); ok {
			return err
		}

		caught := false
		throwSig, isThrow := err.(*ThrowSignal)

		for _, c := range n.Catches {
			if c.Type == "" {
				e.stash.Set("error", err.Error())
				if isThrow {
					e.stash.Set("error", map[string]interface{}{
						"type": throwSig.Type,
						"info": throwSig.Message,
					})
				}
				if execErr := e.execNodes(c.Body); execErr != nil {
					return execErr
				}
				caught = true
				break
			}
			if isThrow && strings.HasPrefix(throwSig.Type, c.Type) {
				e.stash.Set("error", map[string]interface{}{
					"type": throwSig.Type,
					"info": throwSig.Message,
				})
				if execErr := e.execNodes(c.Body); execErr != nil {
					return execErr
				}
				caught = true
				break
			}
		}

		if !caught {
			return err
		}
	}

	if n.Final != nil {
		if err := e.execNodes(n.Final); err != nil {
			return err
		}
	}
	return nil
}

func (e *Evaluator) execThrow(n *ThrowNode) error {
	typeVal, err := e.evalExpr(n.Type)
	if err != nil {
		return err
	}
	var msg string
	if n.Message != nil {
		msgVal, err := e.evalExpr(n.Message)
		if err != nil {
			return err
		}
		msg = toString(msgVal)
	}
	return &ThrowSignal{Type: toString(typeVal), Message: msg}
}

// resolveTemplateName extracts a template name from an expression.
// For bare identifiers like `header`, uses the name literally.
// For dotted paths like `header.tt`, joins with dots (filenames).
// For strings, evaluates normally.
func (e *Evaluator) resolveTemplateName(expr Expr) string {
	switch ex := expr.(type) {
	case *IdentExpr:
		parts := make([]string, len(ex.Segments))
		for i, seg := range ex.Segments {
			parts[i] = seg.Name
		}
		return strings.Join(parts, ".")
	case *StringExpr:
		val, _ := e.evalString(ex)
		return toString(val)
	default:
		val, _ := e.evalExpr(expr)
		return toString(val)
	}
}

// Expression evaluation

func (e *Evaluator) evalExpr(expr Expr) (interface{}, error) {
	switch ex := expr.(type) {
	case *IdentExpr:
		return e.evalIdent(ex)
	case *StringExpr:
		return e.evalString(ex)
	case *NumberExpr:
		return e.evalNumber(ex)
	case *ArrayExpr:
		return e.evalArray(ex)
	case *HashExpr:
		return e.evalHash(ex)
	case *BinOpExpr:
		return e.evalBinOp(ex)
	case *UnaryExpr:
		return e.evalUnary(ex)
	case *TernaryExpr:
		return e.evalTernary(ex)
	case *FilterExpr:
		return e.evalFilter(ex)
	case *RangeExpr:
		return e.evalRange(ex)
	default:
		return nil, newEvalError(fmt.Sprintf("unknown expression type: %T", expr))
	}
}

func (e *Evaluator) evalIdent(ident *IdentExpr) (interface{}, error) {
	if len(ident.Segments) == 1 && ident.Segments[0].Name == "loop" && len(ident.Segments[0].Args) == 0 {
		val, _ := e.stash.Get("loop")
		return val, nil
	}

	// Check if it's a macro call
	if len(ident.Segments) == 1 {
		name := ident.Segments[0].Name
		if macro, ok := e.macros[name]; ok {
			return e.callMacro(macro, ident.Segments[0].Args)
		}
	}

	return e.stash.Resolve(ident.Segments, func(exprs []Expr) ([]interface{}, error) {
		return e.evalArgExprs(exprs)
	})
}

func (e *Evaluator) callMacro(macro *MacroNode, argExprs []Expr) (interface{}, error) {
	args, err := e.evalArgExprs(argExprs)
	if err != nil {
		return nil, err
	}

	oldStash := e.stash
	e.stash = oldStash.Clone()
	defer func() { e.stash = oldStash }()

	for i, name := range macro.Args {
		if i < len(args) {
			e.stash.Set(name, args[i])
		} else {
			e.stash.Set(name, nil)
		}
	}

	var buf strings.Builder
	oldOutput := e.output
	e.output = &buf
	err = e.execNodes(macro.Body)
	e.output = oldOutput
	if err != nil {
		if _, ok := err.(*FlowSignal); ok && err.(*FlowSignal).Action == "RETURN" {
			return buf.String(), nil
		}
		return nil, err
	}
	return buf.String(), nil
}

func (e *Evaluator) evalString(s *StringExpr) (interface{}, error) {
	val := s.Value
	if strings.Contains(val, "$") {
		val = e.interpolateString(val)
	}
	return val, nil
}

func (e *Evaluator) interpolateString(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}|\$([a-zA-Z_]\w*(?:\.\w+)*)`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		var varName string
		if strings.HasPrefix(match, "${") {
			varName = match[2 : len(match)-1]
		} else {
			varName = match[1:]
		}
		parts := strings.Split(varName, ".")
		segments := make([]IdentSegment, len(parts))
		for i, p := range parts {
			segments[i] = IdentSegment{Name: p}
		}
		val, err := e.stash.Resolve(segments, nil)
		if err != nil || val == nil {
			return ""
		}
		return toString(val)
	})
}

func (e *Evaluator) evalNumber(n *NumberExpr) (interface{}, error) {
	if n.IsFloat {
		var f float64
		fmt.Sscanf(n.Value, "%f", &f)
		return f, nil
	}
	var i int64
	fmt.Sscanf(n.Value, "%d", &i)
	return int(i), nil
}

func (e *Evaluator) evalArray(a *ArrayExpr) (interface{}, error) {
	result := make([]interface{}, len(a.Elements))
	for i, elem := range a.Elements {
		val, err := e.evalExpr(elem)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

func (e *Evaluator) evalHash(h *HashExpr) (interface{}, error) {
	result := make(map[string]interface{})
	for _, pair := range h.Pairs {
		key, err := e.evalExpr(pair.Key)
		if err != nil {
			return nil, err
		}
		val, err := e.evalExpr(pair.Value)
		if err != nil {
			return nil, err
		}
		result[toString(key)] = val
	}
	return result, nil
}

func (e *Evaluator) evalBinOp(b *BinOpExpr) (interface{}, error) {
	switch b.Op {
	case "and":
		left, err := e.evalExpr(b.Left)
		if err != nil {
			return nil, err
		}
		if !isTruthy(left) {
			return left, nil
		}
		return e.evalExpr(b.Right)

	case "or":
		left, err := e.evalExpr(b.Left)
		if err != nil {
			return nil, err
		}
		if isTruthy(left) {
			return left, nil
		}
		return e.evalExpr(b.Right)
	}

	left, err := e.evalExpr(b.Left)
	if err != nil {
		return nil, err
	}
	right, err := e.evalExpr(b.Right)
	if err != nil {
		return nil, err
	}

	switch b.Op {
	case "_":
		return toString(left) + toString(right), nil

	case "+":
		lf, _ := toFloat(left)
		rf, _ := toFloat(right)
		r := lf + rf
		return autoNumber(r), nil

	case "-":
		lf, _ := toFloat(left)
		rf, _ := toFloat(right)
		r := lf - rf
		return autoNumber(r), nil

	case "*":
		lf, _ := toFloat(left)
		rf, _ := toFloat(right)
		r := lf * rf
		return autoNumber(r), nil

	case "/":
		lf, _ := toFloat(left)
		rf, _ := toFloat(right)
		if rf == 0 {
			return nil, newEvalError("division by zero")
		}
		return lf / rf, nil

	case "%", "mod":
		lf, _ := toFloat(left)
		rf, _ := toFloat(right)
		r := fmod(lf, rf)
		return autoNumber(r), nil

	case "div":
		lf, _ := toFloat(left)
		rf, _ := toFloat(right)
		r := intDiv(lf, rf)
		return autoNumber(r), nil

	case "==":
		return toString(left) == toString(right), nil

	case "!=":
		return toString(left) != toString(right), nil

	case "<":
		lf, le := toFloat(left)
		rf, re := toFloat(right)
		if le == nil && re == nil {
			return lf < rf, nil
		}
		return toString(left) < toString(right), nil

	case ">":
		lf, le := toFloat(left)
		rf, re := toFloat(right)
		if le == nil && re == nil {
			return lf > rf, nil
		}
		return toString(left) > toString(right), nil

	case "<=":
		lf, le := toFloat(left)
		rf, re := toFloat(right)
		if le == nil && re == nil {
			return lf <= rf, nil
		}
		return toString(left) <= toString(right), nil

	case ">=":
		lf, le := toFloat(left)
		rf, re := toFloat(right)
		if le == nil && re == nil {
			return lf >= rf, nil
		}
		return toString(left) >= toString(right), nil

	default:
		return nil, newEvalError(fmt.Sprintf("unknown operator: %s", b.Op))
	}
}

func (e *Evaluator) evalUnary(u *UnaryExpr) (interface{}, error) {
	operand, err := e.evalExpr(u.Operand)
	if err != nil {
		return nil, err
	}

	switch u.Op {
	case "not":
		return !isTruthy(operand), nil
	case "-":
		f, err := toFloat(operand)
		if err != nil {
			return nil, err
		}
		return autoNumber(-f), nil
	default:
		return nil, newEvalError(fmt.Sprintf("unknown unary operator: %s", u.Op))
	}
}

func (e *Evaluator) evalTernary(t *TernaryExpr) (interface{}, error) {
	cond, err := e.evalExpr(t.Condition)
	if err != nil {
		return nil, err
	}
	if isTruthy(cond) {
		return e.evalExpr(t.Then)
	}
	return e.evalExpr(t.Else)
}

func (e *Evaluator) evalFilter(f *FilterExpr) (interface{}, error) {
	input, err := e.evalExpr(f.Input)
	if err != nil {
		return nil, err
	}
	args, err := e.evalArgExprs(f.Args)
	if err != nil {
		return nil, err
	}

	fn, ok := e.filters[f.Name]
	if !ok {
		return nil, newEvalError(fmt.Sprintf("unknown filter: %s", f.Name))
	}
	return fn(toString(input), args), nil
}

func (e *Evaluator) evalRange(r *RangeExpr) (interface{}, error) {
	startVal, err := e.evalExpr(r.Start)
	if err != nil {
		return nil, err
	}
	endVal, err := e.evalExpr(r.End)
	if err != nil {
		return nil, err
	}
	start := toInt(startVal)
	end := toInt(endVal)
	return makeRange(start, end), nil
}

func (e *Evaluator) evalArgExprs(exprs []Expr) ([]interface{}, error) {
	args := make([]interface{}, len(exprs))
	for i, expr := range exprs {
		val, err := e.evalExpr(expr)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}
	return args, nil
}

// autoNumber returns int if the float is a whole number, otherwise float.
func autoNumber(f float64) interface{} {
	if f == math.Trunc(f) && !math.IsInf(f, 0) && f >= math.MinInt64 && f <= math.MaxInt64 {
		return int(f)
	}
	return f
}
