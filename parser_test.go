package tt

import (
	"testing"
)

func TestParseSimpleGet(t *testing.T) {
	tmpl, err := Parse("[% foo %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(tmpl.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tmpl.Children))
	}
	get, ok := tmpl.Children[0].(*GetNode)
	if !ok {
		t.Fatalf("expected GetNode, got %T", tmpl.Children[0])
	}
	ident, ok := get.Expr.(*IdentExpr)
	if !ok {
		t.Fatalf("expected IdentExpr, got %T", get.Expr)
	}
	if ident.Segments[0].Name != "foo" {
		t.Errorf("expected 'foo', got %q", ident.Segments[0].Name)
	}
}

func TestParseDotNotation(t *testing.T) {
	tmpl, err := Parse("[% foo.bar.baz %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	get := tmpl.Children[0].(*GetNode)
	ident := get.Expr.(*IdentExpr)
	if len(ident.Segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(ident.Segments))
	}
	names := []string{"foo", "bar", "baz"}
	for i, seg := range ident.Segments {
		if seg.Name != names[i] {
			t.Errorf("segment %d: expected %q, got %q", i, names[i], seg.Name)
		}
	}
}

func TestParseSet(t *testing.T) {
	tmpl, err := Parse("[% SET x = 42 %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	set, ok := tmpl.Children[0].(*SetNode)
	if !ok {
		t.Fatalf("expected SetNode, got %T", tmpl.Children[0])
	}
	if len(set.Pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(set.Pairs))
	}
}

func TestParseImplicitSet(t *testing.T) {
	tmpl, err := Parse("[% x = 42 %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := tmpl.Children[0].(*SetNode)
	if !ok {
		t.Fatalf("expected SetNode, got %T", tmpl.Children[0])
	}
}

func TestParseIfElse(t *testing.T) {
	tmpl, err := Parse("[% IF x %]yes[% ELSE %]no[% END %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	ifNode, ok := tmpl.Children[0].(*IfNode)
	if !ok {
		t.Fatalf("expected IfNode, got %T", tmpl.Children[0])
	}
	if len(ifNode.Branches) != 1 {
		t.Errorf("expected 1 branch, got %d", len(ifNode.Branches))
	}
	if len(ifNode.Else) != 1 {
		t.Errorf("expected 1 else node, got %d", len(ifNode.Else))
	}
}

func TestParseIfElsif(t *testing.T) {
	tmpl, err := Parse("[% IF a %]1[% ELSIF b %]2[% ELSIF c %]3[% ELSE %]4[% END %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	ifNode := tmpl.Children[0].(*IfNode)
	if len(ifNode.Branches) != 3 {
		t.Errorf("expected 3 branches, got %d", len(ifNode.Branches))
	}
	if len(ifNode.Else) != 1 {
		t.Errorf("expected 1 else node, got %d", len(ifNode.Else))
	}
}

func TestParseForeach(t *testing.T) {
	tmpl, err := Parse("[% FOREACH item IN list %][% item %][% END %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	foreach, ok := tmpl.Children[0].(*ForeachNode)
	if !ok {
		t.Fatalf("expected ForeachNode, got %T", tmpl.Children[0])
	}
	if foreach.LoopVar != "item" {
		t.Errorf("expected loop var 'item', got %q", foreach.LoopVar)
	}
}

func TestParseBlock(t *testing.T) {
	tmpl, err := Parse("[% BLOCK header %]<h1>Hi</h1>[% END %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	block, ok := tmpl.Children[0].(*BlockDefNode)
	if !ok {
		t.Fatalf("expected BlockDefNode, got %T", tmpl.Children[0])
	}
	if block.Name != "header" {
		t.Errorf("expected block name 'header', got %q", block.Name)
	}
}

func TestParseFilter(t *testing.T) {
	tmpl, err := Parse("[% name | upper %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	get, ok := tmpl.Children[0].(*GetNode)
	if !ok {
		t.Fatalf("expected GetNode, got %T", tmpl.Children[0])
	}
	filter, ok := get.Expr.(*FilterExpr)
	if !ok {
		t.Fatalf("expected FilterExpr, got %T", get.Expr)
	}
	if filter.Name != "upper" {
		t.Errorf("expected filter 'upper', got %q", filter.Name)
	}
}

func TestParseExpression(t *testing.T) {
	tmpl, err := Parse("[% a + b * c %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	get := tmpl.Children[0].(*GetNode)
	binop, ok := get.Expr.(*BinOpExpr)
	if !ok {
		t.Fatalf("expected BinOpExpr, got %T", get.Expr)
	}
	if binop.Op != "+" {
		t.Errorf("expected top-level op '+', got %q", binop.Op)
	}
	_, ok = binop.Right.(*BinOpExpr)
	if !ok {
		t.Fatalf("expected right side to be BinOpExpr for *, got %T", binop.Right)
	}
}

func TestParseTernary(t *testing.T) {
	tmpl, err := Parse("[% x ? 'yes' : 'no' %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	get := tmpl.Children[0].(*GetNode)
	ternary, ok := get.Expr.(*TernaryExpr)
	if !ok {
		t.Fatalf("expected TernaryExpr, got %T", get.Expr)
	}
	_ = ternary
}

func TestParseArrayLiteral(t *testing.T) {
	tmpl, err := Parse("[% x = [1, 2, 3] %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	set := tmpl.Children[0].(*SetNode)
	arr, ok := set.Pairs[0].Value.(*ArrayExpr)
	if !ok {
		t.Fatalf("expected ArrayExpr, got %T", set.Pairs[0].Value)
	}
	if len(arr.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestParseHashLiteral(t *testing.T) {
	tmpl, err := Parse("[% x = { a => 1, b => 2 } %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	set := tmpl.Children[0].(*SetNode)
	hash, ok := set.Pairs[0].Value.(*HashExpr)
	if !ok {
		t.Fatalf("expected HashExpr, got %T", set.Pairs[0].Value)
	}
	if len(hash.Pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(hash.Pairs))
	}
}

func TestParseInclude(t *testing.T) {
	tmpl, err := Parse("[% INCLUDE header title='Hello' %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	inc, ok := tmpl.Children[0].(*IncludeNode)
	if !ok {
		t.Fatalf("expected IncludeNode, got %T", tmpl.Children[0])
	}
	if !inc.Localize {
		t.Error("expected INCLUDE to localize")
	}
	if len(inc.Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(inc.Params))
	}
}

func TestParseProcess(t *testing.T) {
	tmpl, err := Parse("[% PROCESS footer %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	inc, ok := tmpl.Children[0].(*IncludeNode)
	if !ok {
		t.Fatalf("expected IncludeNode, got %T", tmpl.Children[0])
	}
	if inc.Localize {
		t.Error("expected PROCESS to not localize")
	}
}

func TestParseSwitch(t *testing.T) {
	input := "[% SWITCH type %][% CASE 'a' %]alpha[% CASE 'b' %]beta[% CASE %]other[% END %]"
	tmpl, err := Parse(input, "", "")
	if err != nil {
		t.Fatal(err)
	}
	sw, ok := tmpl.Children[0].(*SwitchNode)
	if !ok {
		t.Fatalf("expected SwitchNode, got %T", tmpl.Children[0])
	}
	if len(sw.Cases) != 3 {
		t.Errorf("expected 3 cases, got %d", len(sw.Cases))
	}
	if len(sw.Cases[2].Values) != 0 {
		t.Error("expected default case (no values)")
	}
}

func TestParseRange(t *testing.T) {
	tmpl, err := Parse("[% FOREACH i IN 1..10 %][% i %][% END %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	foreach := tmpl.Children[0].(*ForeachNode)
	_, ok := foreach.List.(*RangeExpr)
	if !ok {
		t.Fatalf("expected RangeExpr, got %T", foreach.List)
	}
}

func TestParseFilterBlock(t *testing.T) {
	tmpl, err := Parse("[% FILTER upper %]hello[% END %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	fb, ok := tmpl.Children[0].(*FilterBlockNode)
	if !ok {
		t.Fatalf("expected FilterBlockNode, got %T", tmpl.Children[0])
	}
	if fb.Name != "upper" {
		t.Errorf("expected filter name 'upper', got %q", fb.Name)
	}
}

func TestParseTryCatch(t *testing.T) {
	input := "[% TRY %]risky[% CATCH %]error[% END %]"
	tmpl, err := Parse(input, "", "")
	if err != nil {
		t.Fatal(err)
	}
	tryNode, ok := tmpl.Children[0].(*TryNode)
	if !ok {
		t.Fatalf("expected TryNode, got %T", tmpl.Children[0])
	}
	if len(tryNode.Catches) != 1 {
		t.Errorf("expected 1 catch, got %d", len(tryNode.Catches))
	}
}

func TestParseMethodCall(t *testing.T) {
	tmpl, err := Parse("[% foo.bar(1, 2) %]", "", "")
	if err != nil {
		t.Fatal(err)
	}
	get := tmpl.Children[0].(*GetNode)
	ident := get.Expr.(*IdentExpr)
	if len(ident.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(ident.Segments))
	}
	if len(ident.Segments[1].Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(ident.Segments[1].Args))
	}
}
