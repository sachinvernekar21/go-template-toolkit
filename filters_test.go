package tt

import (
	"strings"
	"testing"
)

func TestFilterHTML(t *testing.T) {
	result := filterHTML(`<b>"Hello" & 'World'</b>`, nil)
	expected := `&lt;b&gt;&quot;Hello&quot; &amp; 'World'&lt;/b&gt;`
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFilterXML(t *testing.T) {
	result := filterXML(`<b>"Hello" & 'World'</b>`, nil)
	expected := `&lt;b&gt;&quot;Hello&quot; &amp; &apos;World&apos;&lt;/b&gt;`
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFilterURI(t *testing.T) {
	result := filterURI("hello world&foo=bar", nil)
	expected := "hello+world%26foo%3Dbar"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFilterUpper(t *testing.T) {
	if filterUpper("hello", nil) != "HELLO" {
		t.Error("upper filter failed")
	}
}

func TestFilterLower(t *testing.T) {
	if filterLower("HELLO", nil) != "hello" {
		t.Error("lower filter failed")
	}
}

func TestFilterTrim(t *testing.T) {
	if filterTrim("  hello  ", nil) != "hello" {
		t.Error("trim filter failed")
	}
}

func TestFilterCollapse(t *testing.T) {
	result := filterCollapse("  hello   world  ", nil)
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestFilterIndent(t *testing.T) {
	result := filterIndent("hello\nworld", []interface{}{2})
	expected := "  hello\n  world"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFilterTruncate(t *testing.T) {
	result := filterTruncate("Hello, World!", []interface{}{5})
	if result != "Hello..." {
		t.Errorf("expected 'Hello...', got %q", result)
	}
}

func TestFilterRepeat(t *testing.T) {
	result := filterRepeat("ab", []interface{}{3})
	if result != "ababab" {
		t.Errorf("expected 'ababab', got %q", result)
	}
}

func TestFilterReplace(t *testing.T) {
	result := filterReplace("foo bar foo", []interface{}{"foo", "baz"})
	if result != "baz bar baz" {
		t.Errorf("expected 'baz bar baz', got %q", result)
	}
}

func TestFilterRemove(t *testing.T) {
	result := filterRemove("Hello123World", []interface{}{"[0-9]+"})
	if result != "HelloWorld" {
		t.Errorf("expected 'HelloWorld', got %q", result)
	}
}

func TestFilterFormat(t *testing.T) {
	result := filterFormat("hello\nworld", []interface{}{"[%s]"})
	if result != "[hello]\n[world]" {
		t.Errorf("expected '[hello]\\n[world]', got %q", result)
	}
}

func TestFilterNull(t *testing.T) {
	if filterNull("anything", nil) != "" {
		t.Error("null filter should return empty")
	}
}

func TestFilterUcfirst(t *testing.T) {
	if filterUcfirst("hello", nil) != "Hello" {
		t.Error("ucfirst filter failed")
	}
}

func TestFilterLcfirst(t *testing.T) {
	if filterLcfirst("Hello", nil) != "hello" {
		t.Error("lcfirst filter failed")
	}
}

func TestFilterHTMLLineBreak(t *testing.T) {
	result := filterHTMLLineBreak("hello\nworld", nil)
	if result != "hello<br>\nworld" {
		t.Errorf("expected 'hello<br>\\nworld', got %q", result)
	}
}

func TestFilterHTMLPara(t *testing.T) {
	input := "First paragraph.\n\nSecond paragraph.\n\nThird."
	result := filterHTMLPara(input, nil)
	if strings.Count(result, "<p>") != 3 {
		t.Errorf("expected 3 <p> tags, got %d in %q", strings.Count(result, "<p>"), result)
	}
	if !strings.Contains(result, "First paragraph.") {
		t.Error("expected first paragraph content")
	}
	if !strings.Contains(result, "Second paragraph.") {
		t.Error("expected second paragraph content")
	}
}

func TestFilterHTMLBreak(t *testing.T) {
	input := "line one\n\nline two\n\nline three"
	result := filterHTMLBreak(input, nil)
	expected := "line one<br>\n<br>\nline two<br>\n<br>\nline three"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFilterURL(t *testing.T) {
	result := filterURL("hello world&foo=bar", nil)
	expected := "hello+world%26foo%3Dbar"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
