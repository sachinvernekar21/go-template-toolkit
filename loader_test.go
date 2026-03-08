package tt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStringLoader(t *testing.T) {
	loader := NewStringLoader(map[string]string{
		"header": "<h1>Hello</h1>",
		"footer": "<footer>Bye</footer>",
	})

	content, err := loader.Load("header")
	if err != nil {
		t.Fatal(err)
	}
	if content != "<h1>Hello</h1>" {
		t.Errorf("unexpected content: %q", content)
	}

	_, err = loader.Load("missing")
	if err == nil {
		t.Error("expected error for missing template")
	}
}

func TestFileLoader(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "test.tt"), []byte("Hello [% name %]"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	loader := NewFileLoader([]string{tmpDir})

	content, err := loader.Load("test.tt")
	if err != nil {
		t.Fatal(err)
	}
	if content != "Hello [% name %]" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestFileLoaderCache(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cached.tt")

	err := os.WriteFile(path, []byte("original"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	loader := NewFileLoader([]string{tmpDir})
	loader.Cache = true

	content, _ := loader.Load("cached.tt")
	if content != "original" {
		t.Errorf("expected 'original', got %q", content)
	}

	os.WriteFile(path, []byte("modified"), 0644)

	content, _ = loader.Load("cached.tt")
	if content != "original" {
		t.Error("expected cached value, but got modified")
	}

	loader.ClearCache()
	content, _ = loader.Load("cached.tt")
	if content != "modified" {
		t.Error("expected new value after cache clear")
	}
}

func TestFileLoaderMultiplePaths(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir2, "only_in_dir2.tt"), []byte("from dir2"), 0644)

	loader := NewFileLoader([]string{dir1, dir2})

	content, err := loader.Load("only_in_dir2.tt")
	if err != nil {
		t.Fatal(err)
	}
	if content != "from dir2" {
		t.Errorf("expected 'from dir2', got %q", content)
	}
}

func TestFileLoaderAbsolute(t *testing.T) {
	tmpDir := t.TempDir()
	absPath := filepath.Join(tmpDir, "abs.tt")
	os.WriteFile(absPath, []byte("absolute content"), 0644)

	loader := NewFileLoader(nil)
	_, err := loader.Load(absPath)
	if err == nil {
		t.Error("expected error for absolute path when ABSOLUTE is disabled")
	}

	loader.Absolute = true
	content, err := loader.Load(absPath)
	if err != nil {
		t.Fatal(err)
	}
	if content != "absolute content" {
		t.Errorf("expected 'absolute content', got %q", content)
	}
}

func TestFileLoaderMissing(t *testing.T) {
	loader := NewFileLoader([]string{"/nonexistent"})
	_, err := loader.Load("missing.tt")
	if err == nil {
		t.Error("expected error for missing template")
	}
}
