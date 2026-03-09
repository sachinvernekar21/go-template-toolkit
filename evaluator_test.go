package tt

import (
	"strings"
	"testing"
)

func evalTemplate(t *testing.T, tmpl string, vars map[string]interface{}) string {
	t.Helper()
	engine := New()
	result, err := engine.ProcessString(tmpl, vars)
	if err != nil {
		t.Fatalf("error evaluating template: %v", err)
	}
	return result
}

func TestEvalSimpleVar(t *testing.T) {
	result := evalTemplate(t, "Hello [% name %]!", map[string]interface{}{"name": "World"})
	if result != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", result)
	}
}

func TestEvalDotNotation(t *testing.T) {
	vars := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
		},
	}
	result := evalTemplate(t, "[% user.name %]", vars)
	if result != "Alice" {
		t.Errorf("expected 'Alice', got %q", result)
	}
}

func TestEvalSet(t *testing.T) {
	result := evalTemplate(t, "[% SET x = 'hello' %][% x %]", nil)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestEvalImplicitSet(t *testing.T) {
	result := evalTemplate(t, "[% x = 42 %][% x %]", nil)
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
	}
}

func TestEvalDefault(t *testing.T) {
	result := evalTemplate(t, "[% DEFAULT title = 'Untitled' %][% title %]", nil)
	if result != "Untitled" {
		t.Errorf("expected 'Untitled', got %q", result)
	}
	result = evalTemplate(t, "[% DEFAULT title = 'Untitled' %][% title %]", map[string]interface{}{"title": "My Page"})
	if result != "My Page" {
		t.Errorf("expected 'My Page', got %q", result)
	}
}

func TestEvalIfElse(t *testing.T) {
	tmpl := "[% IF show %]yes[% ELSE %]no[% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{"show": true})
	if result != "yes" {
		t.Errorf("expected 'yes', got %q", result)
	}
	result = evalTemplate(t, tmpl, map[string]interface{}{"show": false})
	if result != "no" {
		t.Errorf("expected 'no', got %q", result)
	}
}

func TestEvalUnless(t *testing.T) {
	tmpl := "[% UNLESS hidden %]visible[% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{"hidden": false})
	if result != "visible" {
		t.Errorf("expected 'visible', got %q", result)
	}
	result = evalTemplate(t, tmpl, map[string]interface{}{"hidden": true})
	if result != "" {
		t.Errorf("expected '', got %q", result)
	}
}

func TestEvalElsif(t *testing.T) {
	tmpl := "[% IF x == 1 %]one[% ELSIF x == 2 %]two[% ELSE %]other[% END %]"
	if evalTemplate(t, tmpl, map[string]interface{}{"x": 1}) != "one" {
		t.Error("expected 'one'")
	}
	if evalTemplate(t, tmpl, map[string]interface{}{"x": 2}) != "two" {
		t.Error("expected 'two'")
	}
	if evalTemplate(t, tmpl, map[string]interface{}{"x": 3}) != "other" {
		t.Error("expected 'other'")
	}
}

func TestEvalForeach(t *testing.T) {
	tmpl := "[% FOREACH item IN list %][% item %] [% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "a b c " {
		t.Errorf("expected 'a b c ', got %q", result)
	}
}

func TestEvalForeachLoop(t *testing.T) {
	tmpl := "[% FOREACH item IN list %][% loop.count %]:[% item %] [% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "1:a 2:b 3:c " {
		t.Errorf("expected '1:a 2:b 3:c ', got %q", result)
	}
}

func TestEvalRange(t *testing.T) {
	tmpl := "[% FOREACH i IN 1..5 %][% i %][% END %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "12345" {
		t.Errorf("expected '12345', got %q", result)
	}
}

func TestEvalArithmetic(t *testing.T) {
	tests := []struct {
		tmpl   string
		expect string
	}{
		{"[% 2 + 3 %]", "5"},
		{"[% 10 - 4 %]", "6"},
		{"[% 3 * 4 %]", "12"},
		{"[% 15 / 6 %]", "2.5"},
		{"[% 15 div 6 %]", "2"},
		{"[% 15 mod 6 %]", "3"},
	}
	for _, tc := range tests {
		result := evalTemplate(t, tc.tmpl, nil)
		if result != tc.expect {
			t.Errorf("%s: expected %q, got %q", tc.tmpl, tc.expect, result)
		}
	}
}

func TestEvalConcat(t *testing.T) {
	result := evalTemplate(t, "[% 'hello' _ ' ' _ 'world' %]", nil)
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestEvalTernary(t *testing.T) {
	tmpl := "[% x ? 'yes' : 'no' %]"
	if evalTemplate(t, tmpl, map[string]interface{}{"x": true}) != "yes" {
		t.Error("expected 'yes'")
	}
	if evalTemplate(t, tmpl, map[string]interface{}{"x": false}) != "no" {
		t.Error("expected 'no'")
	}
}

func TestEvalFilter(t *testing.T) {
	result := evalTemplate(t, "[% name | upper %]", map[string]interface{}{"name": "hello"})
	if result != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", result)
	}
}

func TestEvalFilterChain(t *testing.T) {
	result := evalTemplate(t, "[% '  Hello World  ' | trim | upper %]", nil)
	if result != "HELLO WORLD" {
		t.Errorf("expected 'HELLO WORLD', got %q", result)
	}
}

func TestEvalFilterBlock(t *testing.T) {
	result := evalTemplate(t, "[% FILTER upper %]hello world[% END %]", nil)
	if result != "HELLO WORLD" {
		t.Errorf("expected 'HELLO WORLD', got %q", result)
	}
}

func TestEvalHTMLFilter(t *testing.T) {
	result := evalTemplate(t, "[% content | html %]", map[string]interface{}{
		"content": "<b>Hello</b>",
	})
	if result != "&lt;b&gt;Hello&lt;/b&gt;" {
		t.Errorf("expected escaped HTML, got %q", result)
	}
}

func TestEvalBlock(t *testing.T) {
	tmpl := "[% BLOCK greet %]Hello [% name %][% END %][% INCLUDE greet name='World' %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result)
	}
}

func TestEvalSwitch(t *testing.T) {
	tmpl := "[% SWITCH type %][% CASE 'a' %]alpha[% CASE 'b' %]beta[% CASE %]other[% END %]"
	if evalTemplate(t, tmpl, map[string]interface{}{"type": "a"}) != "alpha" {
		t.Error("expected 'alpha'")
	}
	if evalTemplate(t, tmpl, map[string]interface{}{"type": "b"}) != "beta" {
		t.Error("expected 'beta'")
	}
	if evalTemplate(t, tmpl, map[string]interface{}{"type": "c"}) != "other" {
		t.Error("expected 'other'")
	}
}

func TestEvalVMethod(t *testing.T) {
	result := evalTemplate(t, "[% name.length %]", map[string]interface{}{"name": "hello"})
	if result != "5" {
		t.Errorf("expected '5', got %q", result)
	}
}

func TestEvalListVMethod(t *testing.T) {
	result := evalTemplate(t, "[% list.size %]", map[string]interface{}{
		"list": []interface{}{1, 2, 3},
	})
	if result != "3" {
		t.Errorf("expected '3', got %q", result)
	}
}

func TestEvalListJoin(t *testing.T) {
	result := evalTemplate(t, "[% list.join(', ') %]", map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "a, b, c" {
		t.Errorf("expected 'a, b, c', got %q", result)
	}
}

func TestEvalHashVMethod(t *testing.T) {
	result := evalTemplate(t, "[% hash.size %]", map[string]interface{}{
		"hash": map[string]interface{}{"a": 1, "b": 2},
	})
	if result != "2" {
		t.Errorf("expected '2', got %q", result)
	}
}

func TestEvalStringInterpolation(t *testing.T) {
	result := evalTemplate(t, `[% SET name = 'World' %][% "Hello $name" %]`, nil)
	if result != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result)
	}
}

func TestEvalStringInterpolationBraces(t *testing.T) {
	result := evalTemplate(t, `[% SET name = 'World' %][% "Hello ${name}!" %]`, nil)
	if result != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", result)
	}
}

func TestEvalCall(t *testing.T) {
	called := false
	vars := map[string]interface{}{
		"counter": func() string {
			called = true
			return "ignored"
		},
	}
	engine := New()
	engine.ProcessString("[% CALL counter %]", vars)
	// CALL shouldn't error, just evaluated
	_ = called
}

func TestEvalIncludeWithLoader(t *testing.T) {
	engine := New()
	engine.SetLoader(NewStringLoader(map[string]string{
		"header": "==[% title %]==",
	}))
	result, err := engine.ProcessString("[% INCLUDE header title='Test' %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "==Test==" {
		t.Errorf("expected '==Test==', got %q", result)
	}
}

func TestEvalProcessWithLoader(t *testing.T) {
	engine := New()
	engine.SetLoader(NewStringLoader(map[string]string{
		"footer": "--[% name %]--",
	}))
	result, err := engine.ProcessString("[% SET name = 'global' %][% PROCESS footer %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "--global--" {
		t.Errorf("expected '--global--', got %q", result)
	}
}

func TestEvalMacro(t *testing.T) {
	tmpl := "[% MACRO greet(name) BLOCK %]Hello [% name %][% END %][% greet('World') %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result)
	}
}

func TestEvalTryCatch(t *testing.T) {
	tmpl := "[% TRY %][% THROW 'oops' 'bad thing' %][% CATCH %]caught[% END %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "caught" {
		t.Errorf("expected 'caught', got %q", result)
	}
}

func TestEvalWhile(t *testing.T) {
	tmpl := "[% SET i = 0 %][% WHILE i < 5 %][% i %][% i = i + 1 %][% END %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "01234" {
		t.Errorf("expected '01234', got %q", result)
	}
}

func TestEvalNot(t *testing.T) {
	result := evalTemplate(t, "[% IF NOT false %]yes[% END %]", map[string]interface{}{"false": false})
	if result != "yes" {
		t.Errorf("expected 'yes', got %q", result)
	}
}

func TestEvalComparisonOps(t *testing.T) {
	tests := []struct {
		tmpl   string
		vars   map[string]interface{}
		expect string
	}{
		{"[% IF a > b %]yes[% END %]", map[string]interface{}{"a": 5, "b": 3}, "yes"},
		{"[% IF a < b %]yes[% END %]", map[string]interface{}{"a": 3, "b": 5}, "yes"},
		{"[% IF a >= b %]yes[% END %]", map[string]interface{}{"a": 5, "b": 5}, "yes"},
		{"[% IF a <= b %]yes[% END %]", map[string]interface{}{"a": 5, "b": 5}, "yes"},
		{"[% IF a != b %]yes[% END %]", map[string]interface{}{"a": 1, "b": 2}, "yes"},
	}
	for _, tc := range tests {
		result := evalTemplate(t, tc.tmpl, tc.vars)
		if result != tc.expect {
			t.Errorf("%s: expected %q, got %q", tc.tmpl, tc.expect, result)
		}
	}
}

func TestEvalAndOr(t *testing.T) {
	result := evalTemplate(t, "[% IF a and b %]yes[% END %]", map[string]interface{}{"a": true, "b": true})
	if result != "yes" {
		t.Error("expected 'yes' for a and b")
	}
	result = evalTemplate(t, "[% IF a or b %]yes[% END %]", map[string]interface{}{"a": false, "b": true})
	if result != "yes" {
		t.Error("expected 'yes' for a or b")
	}
}

func TestEvalArrayLiteral(t *testing.T) {
	tmpl := "[% SET items = ['a', 'b', 'c'] %][% FOREACH i IN items %][% i %][% END %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "abc" {
		t.Errorf("expected 'abc', got %q", result)
	}
}

func TestEvalHashLiteral(t *testing.T) {
	tmpl := "[% SET person = { name => 'Alice', age => 30 } %][% person.name %] is [% person.age %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "Alice is 30" {
		t.Errorf("expected 'Alice is 30', got %q", result)
	}
}

func TestEvalCustomFilter(t *testing.T) {
	engine := New()
	engine.AddFilter("reverse", func(input string, args []interface{}) string {
		runes := []rune(input)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})
	result, err := engine.ProcessString("[% 'hello' | reverse %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "olleh" {
		t.Errorf("expected 'olleh', got %q", result)
	}
}

func TestEvalChomp(t *testing.T) {
	result := evalTemplate(t, "Hello   \n  [%- name -%]   \nWorld", map[string]interface{}{"name": "X"})
	if result != "HelloXWorld" {
		t.Errorf("expected 'HelloXWorld', got %q", result)
	}
}

func TestEvalForeachNext(t *testing.T) {
	tmpl := "[% FOREACH i IN [1,2,3,4,5] %][% IF i == 3 %][% NEXT %][% END %][% i %][% END %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "1245" {
		t.Errorf("expected '1245', got %q", result)
	}
}

func TestEvalForeachLast(t *testing.T) {
	tmpl := "[% FOREACH i IN [1,2,3,4,5] %][% IF i == 3 %][% LAST %][% END %][% i %][% END %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "12" {
		t.Errorf("expected '12', got %q", result)
	}
}

func TestEvalWrapperWithLoader(t *testing.T) {
	engine := New()
	engine.SetLoader(NewStringLoader(map[string]string{
		"layout": "==[% content %]==",
	}))
	result, err := engine.ProcessString("[% WRAPPER layout %]inner[% END %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "==inner==" {
		t.Errorf("expected '==inner==', got %q", result)
	}
}

func TestEvalNested(t *testing.T) {
	tmpl := `[% FOREACH row IN table %][% FOREACH cell IN row %][% cell %] [% END %]
[% END %]`
	vars := map[string]interface{}{
		"table": []interface{}{
			[]interface{}{"a", "b"},
			[]interface{}{"c", "d"},
		},
	}
	result := evalTemplate(t, tmpl, vars)
	if !strings.Contains(result, "a b") || !strings.Contains(result, "c d") {
		t.Errorf("unexpected nested foreach result: %q", result)
	}
}

func TestEvalScopedForeach(t *testing.T) {
	tmpl := "[% SET x = 'outer' %][% FOREACH i IN [1] %][% SET x = 'inner' %][% END %][% x %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "outer" {
		t.Errorf("expected 'outer' (foreach should scope), got %q", result)
	}
}

func TestEvalFilterWithArgs(t *testing.T) {
	result := evalTemplate(t, "[% 'hello world' | truncate(5) %]", nil)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got %q", result)
	}
}

func TestEvalStruct(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	vars := map[string]interface{}{
		"user": User{Name: "Bob", Age: 25},
	}
	result := evalTemplate(t, "[% user.Name %] is [% user.Age %]", vars)
	if result != "Bob is 25" {
		t.Errorf("expected 'Bob is 25', got %q", result)
	}
}

type testPerson struct {
	Name string
	Age  int
}

func (p testPerson) Greeting() string {
	return "Hello, " + p.Name
}

func TestEvalForeachLoopIndex(t *testing.T) {
	tmpl := "[% FOREACH item IN list %][% loop.index %] [% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "0 1 2 " {
		t.Errorf("expected '0 1 2 ', got %q", result)
	}
}

func TestEvalForeachLoopFirst(t *testing.T) {
	tmpl := "[% FOREACH item IN list %][% IF loop.first %]FIRST[% END %][% item %] [% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "FIRSTa b c " {
		t.Errorf("expected 'FIRSTa b c ', got %q", result)
	}
}

func TestEvalForeachLoopLast(t *testing.T) {
	tmpl := "[% FOREACH item IN list %][% item %][% IF loop.last %]LAST[% END %] [% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "a b cLAST " {
		t.Errorf("expected 'a b cLAST ', got %q", result)
	}
}

func TestEvalForeachLoopSize(t *testing.T) {
	tmpl := "[% FOREACH item IN list %][% loop.size %][% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "333" {
		t.Errorf("expected '333', got %q", result)
	}
}

func TestEvalForeachLoopMax(t *testing.T) {
	tmpl := "[% FOREACH item IN list %][% loop.max %][% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "222" {
		t.Errorf("expected '222', got %q", result)
	}
}

func TestEvalForeachLoopPrev(t *testing.T) {
	tmpl := "[% FOREACH item IN list %]([% loop.prev %])[% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "()(a)(b)" {
		t.Errorf("expected '()(a)(b)', got %q", result)
	}
}

func TestEvalForeachLoopNext(t *testing.T) {
	tmpl := "[% FOREACH item IN list %]([% loop.next %])[% END %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "(b)(c)()" {
		t.Errorf("expected '(b)(c)()', got %q", result)
	}
}

func TestEvalTryCatchFinal(t *testing.T) {
	tmpl := "[% TRY %][% THROW 'err' 'oops' %][% CATCH %]caught[% FINAL %]cleanup[% END %]"
	result := evalTemplate(t, tmpl, nil)
	if result != "caughtcleanup" {
		t.Errorf("expected 'caughtcleanup', got %q", result)
	}

	tmpl = "[% TRY %]ok[% FINAL %]cleanup[% END %]"
	result = evalTemplate(t, tmpl, nil)
	if result != "okcleanup" {
		t.Errorf("expected 'okcleanup', got %q", result)
	}
}

func TestEvalExplicitGet(t *testing.T) {
	result := evalTemplate(t, "[% GET name %]", map[string]interface{}{"name": "Alice"})
	if result != "Alice" {
		t.Errorf("expected 'Alice', got %q", result)
	}
}

func TestEvalStructMethod(t *testing.T) {
	vars := map[string]interface{}{
		"user": &testPerson{Name: "Alice", Age: 30},
	}
	result := evalTemplate(t, "[% user.Greeting %]", vars)
	if result != "Hello, Alice" {
		t.Errorf("expected 'Hello, Alice', got %q", result)
	}
}

// --- Lowercase directive tests ---

func TestEvalLowercaseForeach(t *testing.T) {
	tmpl := "[% foreach item in list %][% item %] [% end %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "a b c " {
		t.Errorf("expected 'a b c ', got %q", result)
	}
}

func TestEvalLowercaseIfElse(t *testing.T) {
	tmpl := "[% if show %]yes[% else %]no[% end %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{"show": true})
	if result != "yes" {
		t.Errorf("expected 'yes', got %q", result)
	}
	result = evalTemplate(t, tmpl, map[string]interface{}{"show": false})
	if result != "no" {
		t.Errorf("expected 'no', got %q", result)
	}
}

func TestEvalLowercaseSet(t *testing.T) {
	result := evalTemplate(t, "[% set x = 'hello' %][% x %]", nil)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestEvalLowercaseMixed(t *testing.T) {
	tmpl := "[% foreach item in list %][% if item == 'b' %]FOUND[% end %][% end %]"
	result := evalTemplate(t, tmpl, map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	})
	if result != "FOUND" {
		t.Errorf("expected 'FOUND', got %q", result)
	}
}

func TestEvalLowercaseUnless(t *testing.T) {
	result := evalTemplate(t, "[% unless hidden %]visible[% end %]", map[string]interface{}{"hidden": false})
	if result != "visible" {
		t.Errorf("expected 'visible', got %q", result)
	}
}

func TestEvalLowercaseBlock(t *testing.T) {
	tmpl := "[% block greet %]Hello [% name %][% end %][% include greet name='World' %]"
	engine := New()
	result, err := engine.ProcessString(tmpl, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result)
	}
}

func TestEvalLowercaseFilter(t *testing.T) {
	result := evalTemplate(t, "[% filter upper %]hello[% end %]", nil)
	if result != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", result)
	}
}

// --- Dynamic $variable key access tests ---

func TestEvalDollarVarAccess(t *testing.T) {
	vars := map[string]interface{}{
		"data": map[string]interface{}{
			"color": "red",
			"size":  "large",
		},
		"key": "color",
	}
	result := evalTemplate(t, "[% data.$key %]", vars)
	if result != "red" {
		t.Errorf("expected 'red', got %q", result)
	}
}

func TestEvalDollarVarInForeach(t *testing.T) {
	vars := map[string]interface{}{
		"attributes": map[string]interface{}{
			"color": map[string]interface{}{"name": "Color", "value": "red"},
			"size":  map[string]interface{}{"name": "Size", "value": "large"},
		},
		"items": []interface{}{"color", "size"},
	}
	tmpl := "[% foreach item in items %][% attributes.$item.name %]=[% attributes.$item.value %] [% end %]"
	result := evalTemplate(t, tmpl, vars)
	if result != "Color=red Size=large " {
		t.Errorf("expected 'Color=red Size=large ', got %q", result)
	}
}

func TestEvalDollarVarNested(t *testing.T) {
	vars := map[string]interface{}{
		"showtimesAttributes": map[string]interface{}{
			"morning": map[string]interface{}{"name": "Morning Show", "time": "9am"},
			"evening": map[string]interface{}{"name": "Evening Show", "time": "7pm"},
		},
		"items": []interface{}{"morning", "evening"},
	}
	tmpl := "[% foreach item in items %][% if showtimesAttributes.$item.name %][% showtimesAttributes.$item.name %] at [% showtimesAttributes.$item.time %]\n[% end %][% end %]"
	result := evalTemplate(t, tmpl, vars)
	if !strings.Contains(result, "Morning Show at 9am") {
		t.Errorf("expected 'Morning Show at 9am' in result, got %q", result)
	}
	if !strings.Contains(result, "Evening Show at 7pm") {
		t.Errorf("expected 'Evening Show at 7pm' in result, got %q", result)
	}
}

func TestEvalDollarVarMissing(t *testing.T) {
	vars := map[string]interface{}{
		"data": map[string]interface{}{"a": 1},
		"key":  "missing",
	}
	result := evalTemplate(t, "[% data.$key %]", vars)
	if result != "" {
		t.Errorf("expected empty for missing key, got %q", result)
	}
}
