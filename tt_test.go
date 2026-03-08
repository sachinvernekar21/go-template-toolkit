package tt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineProcessString(t *testing.T) {
	engine := New()
	result, err := engine.ProcessString("Hello [% name %]!", map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", result)
	}
}

func TestEngineProcess(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.tt"), []byte("Title: [% title %]"), 0644)

	engine := New(Config{IncludePath: []string{tmpDir}})
	var buf strings.Builder
	err := engine.Process("main.tt", map[string]interface{}{"title": "Test"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "Title: Test" {
		t.Errorf("expected 'Title: Test', got %q", buf.String())
	}
}

func TestEngineCustomTags(t *testing.T) {
	engine := New(Config{StartTag: "<%", EndTag: "%>"})
	result, err := engine.ProcessString("Hello <% name %>!", map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", result)
	}
}

func TestEngineAddFilter(t *testing.T) {
	engine := New()
	engine.AddFilter("exclaim", func(input string, args []interface{}) string {
		return input + "!!!"
	})
	result, err := engine.ProcessString("[% 'wow' | exclaim %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "wow!!!" {
		t.Errorf("expected 'wow!!!', got %q", result)
	}
}

func TestEngineSetLoader(t *testing.T) {
	engine := New()
	engine.SetLoader(NewStringLoader(map[string]string{
		"greet": "Hi [% who %]",
	}))
	result, err := engine.ProcessString("[% INCLUDE greet who='there' %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hi there" {
		t.Errorf("expected 'Hi there', got %q", result)
	}
}

func TestEngineIncludeFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "header.tt"), []byte("<h1>[% title %]</h1>"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "page.tt"), []byte("[% INCLUDE header.tt title='Page' %]\n<p>body</p>"), 0644)

	engine := New(Config{IncludePath: []string{tmpDir}})
	var buf strings.Builder
	err := engine.Process("page.tt", nil, &buf)
	if err != nil {
		t.Fatal(err)
	}
	expected := "<h1>Page</h1>\n<p>body</p>"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestEngineComplexTemplate(t *testing.T) {
	tmpl := `[% DEFAULT title = 'My Site' %]
<html>
<head><title>[% title %]</title></head>
<body>
[% IF items.size %]
<ul>
[% FOREACH item IN items %]
<li>[% item.name | html %] - $[% item.price %]</li>
[% END %]
</ul>
[% ELSE %]
<p>No items</p>
[% END %]
</body>
</html>`

	vars := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"name": "Widget <A>", "price": 9.99},
			map[string]interface{}{"name": "Gadget", "price": 19.99},
		},
	}

	engine := New()
	result, err := engine.ProcessString(tmpl, vars)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "My Site") {
		t.Error("expected default title")
	}
	if !strings.Contains(result, "Widget &lt;A&gt;") {
		t.Error("expected HTML-escaped widget name")
	}
	if !strings.Contains(result, "9.99") {
		t.Error("expected price 9.99")
	}
	if strings.Contains(result, "No items") {
		t.Error("should not show 'No items' when items exist")
	}
}

func TestEngineWrapperFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "layout.tt"),
		[]byte("<div class='wrapper'>[% content %]</div>"), 0644)

	engine := New(Config{IncludePath: []string{tmpDir}})
	result, err := engine.ProcessString(
		"[% WRAPPER layout.tt %]inner content[% END %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "<div class='wrapper'>inner content</div>" {
		t.Errorf("unexpected wrapper result: %q", result)
	}
}

func TestEngineInsert(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "raw.txt"),
		[]byte("[% NOT_PROCESSED %]"), 0644)

	engine := New(Config{IncludePath: []string{tmpDir}})
	result, err := engine.ProcessString("[% INSERT raw.txt %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "[% NOT_PROCESSED %]" {
		t.Errorf("INSERT should not process content, got %q", result)
	}
}

func TestEngineTryCatchWithType(t *testing.T) {
	tmpl := `[% TRY %][% THROW 'db' 'connection failed' %][% CATCH db %]DB error: [% error.info %][% CATCH %]Other error[% END %]`
	engine := New()
	result, err := engine.ProcessString(tmpl, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "DB error: connection failed" {
		t.Errorf("expected 'DB error: connection failed', got %q", result)
	}
}

func TestEngineProcessNoLoader(t *testing.T) {
	engine := New()
	var buf strings.Builder
	err := engine.Process("anything.tt", nil, &buf)
	if err == nil {
		t.Error("expected error when processing without loader")
	}
}

func TestEngineNilVars(t *testing.T) {
	engine := New()
	result, err := engine.ProcessString("Hello [% name %]!", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello !" {
		t.Errorf("expected 'Hello !', got %q", result)
	}
}

func TestEngineAddVMethod(t *testing.T) {
	engine := New()
	engine.AddVMethod(VMethodScalar, "wordcount", func(val interface{}, args []interface{}) interface{} {
		words := strings.Fields(fmt.Sprint(val))
		return len(words)
	})
	result, err := engine.ProcessString("[% text.wordcount %]", map[string]interface{}{
		"text": "hello beautiful world",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "3" {
		t.Errorf("expected '3', got %q", result)
	}
}

func TestEngineStructMethod(t *testing.T) {
	engine := New()
	result, err := engine.ProcessString("[% user.Greeting %]", map[string]interface{}{
		"user": &testPerson{Name: "Alice", Age: 30},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello, Alice" {
		t.Errorf("expected 'Hello, Alice', got %q", result)
	}
}

func TestEngineCacheEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "cached.tt"), []byte("v1"), 0644)

	engine := New(Config{
		IncludePath:  []string{tmpDir},
		CacheEnabled: true,
	})

	var buf strings.Builder
	err := engine.Process("cached.tt", nil, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "v1" {
		t.Errorf("expected 'v1', got %q", buf.String())
	}

	os.WriteFile(filepath.Join(tmpDir, "cached.tt"), []byte("v2"), 0644)

	buf.Reset()
	err = engine.Process("cached.tt", nil, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "v1" {
		t.Errorf("expected cached 'v1' after file change, got %q", buf.String())
	}
}

func TestEngineRelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "parent.tt"), []byte("parent"), 0644)

	engine := New(Config{IncludePath: []string{subDir}})
	_, err := engine.ProcessString("[% INCLUDE '../parent.tt' %]", nil)
	if err == nil {
		t.Error("expected error for relative path with Relative=false")
	}

	engine = New(Config{IncludePath: []string{subDir}, Relative: true})
	result, err := engine.ProcessString("[% INCLUDE '../parent.tt' %]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "parent" {
		t.Errorf("expected 'parent', got %q", result)
	}
}
