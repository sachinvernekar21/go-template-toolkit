package tt

import (
	"testing"
)

func TestStashGetSet(t *testing.T) {
	s := NewStash(nil)
	s.Set("foo", "bar")
	val, ok := s.Get("foo")
	if !ok || val != "bar" {
		t.Errorf("expected 'bar', got %v", val)
	}
}

func TestStashScope(t *testing.T) {
	parent := NewStash(map[string]interface{}{"x": 1})
	child := parent.Clone()
	child.Set("y", 2)

	val, ok := child.Get("x")
	if !ok || val != 1 {
		t.Errorf("expected child to inherit x=1, got %v", val)
	}

	val, ok = child.Get("y")
	if !ok || val != 2 {
		t.Errorf("expected child to have y=2, got %v", val)
	}

	_, ok = parent.Get("y")
	if ok {
		t.Error("parent should not see child's y")
	}
}

func TestStashDefault(t *testing.T) {
	s := NewStash(map[string]interface{}{"existing": "yes"})
	s.SetDefault("existing", "no")
	s.SetDefault("missing", "default")

	val, _ := s.Get("existing")
	if val != "yes" {
		t.Errorf("expected existing to remain 'yes', got %v", val)
	}

	val, _ = s.Get("missing")
	if val != "default" {
		t.Errorf("expected missing to be 'default', got %v", val)
	}
}

func TestStashResolveMap(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"address": map[string]interface{}{
				"city": "Wonderland",
			},
		},
	}
	s := NewStash(data)

	segments := []IdentSegment{{Name: "user"}, {Name: "name"}}
	val, err := s.Resolve(segments, nil)
	if err != nil {
		t.Fatal(err)
	}
	if val != "Alice" {
		t.Errorf("expected 'Alice', got %v", val)
	}

	segments = []IdentSegment{{Name: "user"}, {Name: "address"}, {Name: "city"}}
	val, err = s.Resolve(segments, nil)
	if err != nil {
		t.Fatal(err)
	}
	if val != "Wonderland" {
		t.Errorf("expected 'Wonderland', got %v", val)
	}
}

type testUser struct {
	Name string
	Age  int
}

func TestStashResolveStruct(t *testing.T) {
	s := NewStash(map[string]interface{}{
		"user": testUser{Name: "Bob", Age: 30},
	})

	segments := []IdentSegment{{Name: "user"}, {Name: "Name"}}
	val, err := s.Resolve(segments, nil)
	if err != nil {
		t.Fatal(err)
	}
	if val != "Bob" {
		t.Errorf("expected 'Bob', got %v", val)
	}
}

func TestStashResolveStructMethod(t *testing.T) {
	s := NewStash(map[string]interface{}{
		"user": &testPerson{Name: "Bob", Age: 25},
	})

	segments := []IdentSegment{{Name: "user"}, {Name: "Greeting"}}
	val, err := s.Resolve(segments, nil)
	if err != nil {
		t.Fatal(err)
	}
	if val != "Hello, Bob" {
		t.Errorf("expected 'Hello, Bob', got %v", val)
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		val    interface{}
		expect bool
	}{
		{nil, false},
		{"", false},
		{"0", false},
		{"hello", true},
		{0, false},
		{1, true},
		{0.0, false},
		{1.5, true},
		{true, true},
		{false, false},
		{[]int{}, false},
		{[]int{1}, true},
	}
	for _, tt := range tests {
		if got := isTruthy(tt.val); got != tt.expect {
			t.Errorf("isTruthy(%v) = %v, want %v", tt.val, got, tt.expect)
		}
	}
}
