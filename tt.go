package tt

import (
	"io"
	"strings"
)

// Config holds configuration options for the template Engine.
type Config struct {
	IncludePath  []string
	StartTag     string // default "[%"
	EndTag       string // default "%]"
	PreChomp     bool
	PostChomp    bool
	CacheEnabled bool
	Absolute     bool
	Relative     bool
}

// Engine is the main entry point for processing templates.
type Engine struct {
	config  Config
	loader  TemplateLoader
	filters map[string]FilterFunc
}

// New creates a new template Engine with the given config.
func New(configs ...Config) *Engine {
	var cfg Config
	if len(configs) > 0 {
		cfg = configs[0]
	}
	if cfg.StartTag == "" {
		cfg.StartTag = "[%"
	}
	if cfg.EndTag == "" {
		cfg.EndTag = "%]"
	}

	var loader TemplateLoader
	if len(cfg.IncludePath) > 0 {
		fl := NewFileLoader(cfg.IncludePath)
		fl.Absolute = cfg.Absolute
		fl.Relative = cfg.Relative
		fl.Cache = cfg.CacheEnabled
		loader = fl
	}

	filters := make(map[string]FilterFunc)
	for k, v := range defaultFilters {
		filters[k] = v
	}

	return &Engine{
		config:  cfg,
		loader:  loader,
		filters: filters,
	}
}

// Process parses and evaluates a named template, writing output to w.
func (e *Engine) Process(name string, vars map[string]interface{}, w io.Writer) error {
	if e.loader == nil {
		return newFileError("no template loader configured (set IncludePath)")
	}

	source, err := e.loader.Load(name)
	if err != nil {
		return err
	}

	return e.processSource(source, vars, w)
}

// ProcessString parses and evaluates a template string, returning the output.
func (e *Engine) ProcessString(tmpl string, vars map[string]interface{}) (string, error) {
	var buf strings.Builder
	err := e.processSource(tmpl, vars, &buf)
	return buf.String(), err
}

func (e *Engine) processSource(source string, vars map[string]interface{}, w io.Writer) error {
	parsed, err := Parse(source, e.config.StartTag, e.config.EndTag)
	if err != nil {
		return err
	}

	stash := NewStash(vars)
	eval := NewEvaluator(stash, w, e.loader)

	for name, fn := range e.filters {
		eval.AddFilter(name, fn)
	}

	return eval.Execute(parsed)
}

// AddFilter registers a custom filter function.
func (e *Engine) AddFilter(name string, fn FilterFunc) {
	e.filters[name] = fn
}

// AddVMethod registers a custom virtual method.
func (e *Engine) AddVMethod(typ VMethodType, name string, fn VMethodFunc) {
	RegisterVMethod(typ, name, fn)
}

// SetLoader sets a custom template loader.
func (e *Engine) SetLoader(loader TemplateLoader) {
	e.loader = loader
}
