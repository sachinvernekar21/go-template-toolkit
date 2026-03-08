package tt

import (
	"strings"
	"testing"
)

func TestScalarVMethods(t *testing.T) {
	tests := []struct {
		name   string
		val    interface{}
		args   []interface{}
		expect interface{}
	}{
		{"length", "hello", nil, 5},
		{"upper", "hello", nil, "HELLO"},
		{"lower", "HELLO", nil, "hello"},
		{"ucfirst", "hello", nil, "Hello"},
		{"lcfirst", "Hello", nil, "hello"},
		{"trim", "  hi  ", nil, "hi"},
		{"collapse", "  hello   world  ", nil, "hello world"},
		{"defined-true", "x", nil, true},
		{"defined-nil", nil, nil, false},
		{"repeat", "ab", []interface{}{3}, "ababab"},
		{"substr-offset", "hello", []interface{}{1}, "ello"},
		{"substr-offset-len", "hello", []interface{}{1, 3}, "ell"},
		{"dquote", "say \"hi\"\nnow", nil, "say \\\"hi\\\"\\nnow"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
		vmethod := tc.name
		if vmethod == "defined-true" || vmethod == "defined-nil" {
			vmethod = "defined"
		}
		if strings.HasPrefix(vmethod, "substr") {
			vmethod = "substr"
		}
			fn, ok := scalarVMethods[vmethod]
			if !ok {
				t.Fatalf("vmethod %q not found", vmethod)
			}
			result := fn(tc.val, tc.args)
			if result != tc.expect {
				t.Errorf("expected %v, got %v", tc.expect, result)
			}
		})
	}
}

func TestScalarSplit(t *testing.T) {
	fn := scalarVMethods["split"]
	result := fn("a,b,c", []interface{}{","})
	list, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected slice, got %T", result)
	}
	if len(list) != 3 || list[0] != "a" || list[1] != "b" || list[2] != "c" {
		t.Errorf("unexpected split result: %v", list)
	}
}

func TestScalarReplace(t *testing.T) {
	fn := scalarVMethods["replace"]
	result := fn("hello world", []interface{}{"world", "Go"})
	if result != "hello Go" {
		t.Errorf("expected 'hello Go', got %v", result)
	}
}

func TestListVMethods(t *testing.T) {
	list := []interface{}{"c", "a", "b"}

	size := listVMethods["size"](list, nil)
	if size != 3 {
		t.Errorf("size: expected 3, got %v", size)
	}

	first := listVMethods["first"](list, nil)
	if first != "c" {
		t.Errorf("first: expected 'c', got %v", first)
	}

	last := listVMethods["last"](list, nil)
	if last != "b" {
		t.Errorf("last: expected 'b', got %v", last)
	}

	joined := listVMethods["join"](list, []interface{}{", "})
	if joined != "c, a, b" {
		t.Errorf("join: expected 'c, a, b', got %v", joined)
	}

	sorted := listVMethods["sort"](list, nil)
	sl := sorted.([]interface{})
	if sl[0] != "a" || sl[1] != "b" || sl[2] != "c" {
		t.Errorf("sort: unexpected result %v", sl)
	}

	reversed := listVMethods["reverse"](list, nil)
	rev := reversed.([]interface{})
	if rev[0] != "b" || rev[1] != "a" || rev[2] != "c" {
		t.Errorf("reverse: unexpected result %v", rev)
	}
}

func TestListUnique(t *testing.T) {
	list := []interface{}{"a", "b", "a", "c", "b"}
	result := listVMethods["unique"](list, nil).([]interface{})
	if len(result) != 3 {
		t.Errorf("unique: expected 3, got %d", len(result))
	}
}

func TestListGrep(t *testing.T) {
	list := []interface{}{"foo", "bar", "foobar", "baz"}
	result := listVMethods["grep"](list, []interface{}{"foo"}).([]interface{})
	if len(result) != 2 {
		t.Errorf("grep: expected 2 matches, got %d", len(result))
	}
}

func TestListSlice(t *testing.T) {
	list := []interface{}{1, 2, 3, 4, 5}
	result := listVMethods["slice"](list, []interface{}{1, 3}).([]interface{})
	if len(result) != 3 || result[0] != 2 || result[2] != 4 {
		t.Errorf("slice: unexpected result %v", result)
	}
}

func TestListMax(t *testing.T) {
	list := []interface{}{1, 2, 3}
	max := listVMethods["max"](list, nil)
	if max != 2 {
		t.Errorf("max: expected 2, got %v", max)
	}
}

func TestHashVMethods(t *testing.T) {
	hash := map[string]interface{}{"name": "Alice", "age": 30}

	size := hashVMethods["size"](hash, nil)
	if size != 2 {
		t.Errorf("size: expected 2, got %v", size)
	}

	exists := hashVMethods["exists"](hash, []interface{}{"name"})
	if exists != true {
		t.Errorf("exists: expected true")
	}

	exists = hashVMethods["exists"](hash, []interface{}{"missing"})
	if exists != false {
		t.Errorf("exists(missing): expected false")
	}

	item := hashVMethods["item"](hash, []interface{}{"name"})
	if item != "Alice" {
		t.Errorf("item: expected 'Alice', got %v", item)
	}
}

func TestHashKeys(t *testing.T) {
	hash := map[string]interface{}{"a": 1, "b": 2}
	keys := hashVMethods["keys"](hash, nil).([]interface{})
	if len(keys) != 2 {
		t.Errorf("keys: expected 2, got %d", len(keys))
	}
}

func TestHashPairs(t *testing.T) {
	hash := map[string]interface{}{"x": 1}
	pairs := hashVMethods["pairs"](hash, nil).([]interface{})
	if len(pairs) != 1 {
		t.Fatalf("pairs: expected 1, got %d", len(pairs))
	}
	pair := pairs[0].(map[string]interface{})
	if pair["key"] != "x" || pair["value"] != 1 {
		t.Errorf("pairs: unexpected result %v", pair)
	}
}

func TestMakeRange(t *testing.T) {
	r := makeRange(1, 5)
	if len(r) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(r))
	}
	if r[0] != 1 || r[4] != 5 {
		t.Errorf("unexpected range: %v", r)
	}

	r = makeRange(5, 1)
	if len(r) != 5 || r[0] != 5 || r[4] != 1 {
		t.Errorf("unexpected reverse range: %v", r)
	}
}

func TestScalarMatch(t *testing.T) {
	fn := scalarVMethods["match"]

	result := fn("hello world", []interface{}{"(\\w+)\\s(\\w+)"})
	matches, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected slice for captures, got %T", result)
	}
	if len(matches) != 2 || matches[0] != "hello" || matches[1] != "world" {
		t.Errorf("unexpected match result: %v", matches)
	}

	result = fn("hello", []interface{}{"he"})
	if result != true {
		t.Errorf("expected true for non-capturing match, got %v", result)
	}

	result = fn("hello", []interface{}{"xyz"})
	if result != nil {
		t.Errorf("expected nil for no match, got %v", result)
	}
}

func TestScalarChunk(t *testing.T) {
	fn := scalarVMethods["chunk"]
	result := fn("abcdefgh", []interface{}{3}).([]interface{})
	if len(result) != 3 || result[0] != "abc" || result[1] != "def" || result[2] != "gh" {
		t.Errorf("unexpected chunk result: %v", result)
	}
}

func TestScalarListVMethod(t *testing.T) {
	fn := scalarVMethods["list"]
	result := fn("hello", nil).([]interface{})
	if len(result) != 1 || result[0] != "hello" {
		t.Errorf("expected [hello], got %v", result)
	}
}

func TestScalarHashVMethod(t *testing.T) {
	fn := scalarVMethods["hash"]
	result := fn("hello", nil).(map[string]interface{})
	if result["value"] != "hello" {
		t.Errorf("expected {value: hello}, got %v", result)
	}
}

func TestListNsort(t *testing.T) {
	list := []interface{}{3, 1, 2, 10}
	result := listVMethods["nsort"](list, nil).([]interface{})
	if result[0] != 1 || result[1] != 2 || result[2] != 3 || result[3] != 10 {
		t.Errorf("nsort: unexpected result %v", result)
	}
}

func TestListPush(t *testing.T) {
	list := []interface{}{"a", "b"}
	result := listVMethods["push"](list, []interface{}{"c"}).([]interface{})
	if len(result) != 3 || result[2] != "c" {
		t.Errorf("push: expected [a b c], got %v", result)
	}
}

func TestListPop(t *testing.T) {
	list := []interface{}{"a", "b", "c"}
	result := listVMethods["pop"](list, nil)
	if result != "c" {
		t.Errorf("pop: expected 'c', got %v", result)
	}
}

func TestListShift(t *testing.T) {
	list := []interface{}{"a", "b", "c"}
	result := listVMethods["shift"](list, nil)
	if result != "a" {
		t.Errorf("shift: expected 'a', got %v", result)
	}
}

func TestListUnshift(t *testing.T) {
	list := []interface{}{"b", "c"}
	result := listVMethods["unshift"](list, []interface{}{"a"}).([]interface{})
	if len(result) != 3 || result[0] != "a" {
		t.Errorf("unshift: expected [a b c], got %v", result)
	}
}

func TestListSplice(t *testing.T) {
	list := []interface{}{"a", "b", "c", "d", "e"}
	removed := listVMethods["splice"](list, []interface{}{1, 2}).([]interface{})
	if len(removed) != 2 || removed[0] != "b" || removed[1] != "c" {
		t.Errorf("splice: expected [b c] removed, got %v", removed)
	}
}

func TestListMerge(t *testing.T) {
	list := []interface{}{"a", "b"}
	other := []interface{}{"c", "d"}
	result := listVMethods["merge"](list, []interface{}{other}).([]interface{})
	if len(result) != 4 || result[0] != "a" || result[3] != "d" {
		t.Errorf("merge: expected [a b c d], got %v", result)
	}
}

func TestListHashVMethod(t *testing.T) {
	list := []interface{}{"a", 1, "b", 2}
	result := listVMethods["hash"](list, nil).(map[string]interface{})
	if result["a"] != 1 || result["b"] != 2 {
		t.Errorf("hash: expected {a:1 b:2}, got %v", result)
	}
}

func TestListDefined(t *testing.T) {
	list := []interface{}{"a", nil, "c"}
	if listVMethods["defined"](list, []interface{}{0}) != true {
		t.Error("defined(0): expected true")
	}
	if listVMethods["defined"](list, []interface{}{1}) != false {
		t.Error("defined(1): expected false for nil element")
	}
	if listVMethods["defined"](list, []interface{}{5}) != false {
		t.Error("defined(5): expected false for out of range")
	}
}

func TestListItem(t *testing.T) {
	list := []interface{}{"a", "b", "c"}
	if listVMethods["item"](list, []interface{}{0}) != "a" {
		t.Error("item(0): expected 'a'")
	}
	if listVMethods["item"](list, []interface{}{2}) != "c" {
		t.Error("item(2): expected 'c'")
	}
	if listVMethods["item"](list, []interface{}{5}) != nil {
		t.Error("item(5): expected nil for out of range")
	}
}

func TestHashValues(t *testing.T) {
	hash := map[string]interface{}{"a": 1, "b": 2}
	result := hashVMethods["values"](hash, nil).([]interface{})
	if len(result) != 2 {
		t.Errorf("values: expected 2, got %d", len(result))
	}
}

func TestHashEach(t *testing.T) {
	hash := map[string]interface{}{"x": 1}
	result := hashVMethods["each"](hash, nil).([]interface{})
	if len(result) != 2 {
		t.Errorf("each: expected 2 elements (key+value), got %d", len(result))
	}
	hasKey, hasVal := false, false
	for _, v := range result {
		if v == "x" {
			hasKey = true
		}
		if v == 1 {
			hasVal = true
		}
	}
	if !hasKey || !hasVal {
		t.Errorf("each: expected key 'x' and value 1, got %v", result)
	}
}

func TestHashListVMethod(t *testing.T) {
	hash := map[string]interface{}{"x": 1}
	result := hashVMethods["list"](hash, nil).([]interface{})
	if len(result) != 1 {
		t.Fatalf("list: expected 1 pair, got %d", len(result))
	}
	pair := result[0].(map[string]interface{})
	if pair["key"] != "x" || pair["value"] != 1 {
		t.Errorf("list: unexpected result %v", pair)
	}
}

func TestHashDefined(t *testing.T) {
	hash := map[string]interface{}{"a": 1, "b": nil}
	if hashVMethods["defined"](hash, []interface{}{"a"}) != true {
		t.Error("defined(a): expected true")
	}
	if hashVMethods["defined"](hash, []interface{}{"b"}) != false {
		t.Error("defined(b): expected false for nil value")
	}
	if hashVMethods["defined"](hash, []interface{}{"c"}) != false {
		t.Error("defined(c): expected false for missing key")
	}
}

func TestHashDelete(t *testing.T) {
	hash := map[string]interface{}{"a": 1, "b": 2, "c": 3}
	hashVMethods["delete"](hash, []interface{}{"b"})
	if _, ok := hash["b"]; ok {
		t.Error("delete: key 'b' should be removed")
	}
	if len(hash) != 2 {
		t.Errorf("delete: expected 2 remaining, got %d", len(hash))
	}
}

func TestHashImport(t *testing.T) {
	hash := map[string]interface{}{"a": 1}
	other := map[string]interface{}{"b": 2, "c": 3}
	hashVMethods["import"](hash, []interface{}{other})
	if hash["b"] != 2 || hash["c"] != 3 {
		t.Errorf("import: expected {a:1 b:2 c:3}, got %v", hash)
	}
}

func TestHashSort(t *testing.T) {
	hash := map[string]interface{}{"c": 3, "a": 1, "b": 2}
	result := hashVMethods["sort"](hash, nil).([]interface{})
	if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("sort: expected [a b c], got %v", result)
	}
}

func TestHashNsort(t *testing.T) {
	hash := map[string]interface{}{"x": 10, "y": 1, "z": 5}
	result := hashVMethods["nsort"](hash, nil).([]interface{})
	if len(result) != 3 || result[0] != "y" || result[1] != "z" || result[2] != "x" {
		t.Errorf("nsort: expected [y z x], got %v", result)
	}
}
