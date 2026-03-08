package tt

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileLoader loads templates from the filesystem using INCLUDE_PATH.
type FileLoader struct {
	IncludePath []string
	Absolute    bool
	Relative    bool
	Cache       bool

	mu    sync.RWMutex
	cache map[string]string
}

func NewFileLoader(paths []string) *FileLoader {
	return &FileLoader{
		IncludePath: paths,
		Cache:       true,
		cache:       make(map[string]string),
	}
}

func (fl *FileLoader) Load(name string) (string, error) {
	if fl.Cache {
		fl.mu.RLock()
		if content, ok := fl.cache[name]; ok {
			fl.mu.RUnlock()
			return content, nil
		}
		fl.mu.RUnlock()
	}

	content, err := fl.loadFromDisk(name)
	if err != nil {
		return "", err
	}

	if fl.Cache {
		fl.mu.Lock()
		fl.cache[name] = content
		fl.mu.Unlock()
	}

	return content, nil
}

func (fl *FileLoader) loadFromDisk(name string) (string, error) {
	if filepath.IsAbs(name) {
		if !fl.Absolute {
			return "", newFileError(fmt.Sprintf("absolute paths not allowed: %s", name))
		}
		return readFile(name)
	}

	if len(name) > 0 && (name[0] == '.' || (len(name) > 1 && name[:2] == "..")) {
		if !fl.Relative {
			return "", newFileError(fmt.Sprintf("relative paths not allowed: %s", name))
		}
	}

	for _, dir := range fl.IncludePath {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return readFile(path)
		}
	}

	return "", newFileError(fmt.Sprintf("template not found: %s", name))
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", newFileError(err.Error())
	}
	return string(data), nil
}

// ClearCache removes all cached templates.
func (fl *FileLoader) ClearCache() {
	fl.mu.Lock()
	fl.cache = make(map[string]string)
	fl.mu.Unlock()
}

// StringLoader loads templates from an in-memory map (useful for testing).
type StringLoader struct {
	Templates map[string]string
}

func NewStringLoader(templates map[string]string) *StringLoader {
	return &StringLoader{Templates: templates}
}

func (sl *StringLoader) Load(name string) (string, error) {
	content, ok := sl.Templates[name]
	if !ok {
		return "", newFileError(fmt.Sprintf("template not found: %s", name))
	}
	return content, nil
}
