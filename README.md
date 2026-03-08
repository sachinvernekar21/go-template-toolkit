# template-toolkit

A Go implementation of Perl's [Template Toolkit](http://template-toolkit.org/) (TT2) template processing system. It provides the familiar `[% ... %]` tag syntax, dot-notation variable access, rich expression evaluation, virtual methods, filters, and template composition directives.

## Installation

```bash
go get github.com/sachinv/template-toolkit
```

## Quick Start

```go
package main

import (
    "fmt"
    "os"

    tt "github.com/sachinv/template-toolkit"
)

func main() {
    engine := tt.New()

    tmpl := `Hello [% name %]! You have [% items.size %] items.
[% FOREACH item IN items %]
  - [% item.name %]: $[% item.price %]
[% END %]`

    vars := map[string]interface{}{
        "name": "Alice",
        "items": []interface{}{
            map[string]interface{}{"name": "Widget", "price": 9.99},
            map[string]interface{}{"name": "Gadget", "price": 19.99},
        },
    }

    result, err := engine.ProcessString(tmpl, vars)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    fmt.Print(result)
}
```

## File-Based Templates

```go
engine := tt.New(tt.Config{
    IncludePath: []string{"./templates", "/shared/templates"},
})

err := engine.Process("page.tt", vars, os.Stdout)
```

## Supported Directives

### Variables

```
[% GET foo %]              # explicit GET (keyword optional)
[% foo %]                  # implicit GET
[% foo.bar.baz %]          # dot notation
[% foo.bar(1, 2) %]        # method calls with args

[% SET x = 42 %]           # explicit SET
[% x = 42 %]               # implicit SET
[% DEFAULT title = 'Hi' %] # set only if undefined
[% CALL some_func() %]     # evaluate without output
```

### Conditionals

```
[% IF condition %]
  ...
[% ELSIF other %]
  ...
[% ELSE %]
  ...
[% END %]

[% UNLESS hidden %]...[% END %]

[% SWITCH type %]
  [% CASE 'a' %]Alpha
  [% CASE 'b' %]Beta
  [% CASE %]Default
[% END %]
```

### Loops

```
[% FOREACH item IN list %]
  [% loop.count %]: [% item %]
[% END %]

[% FOREACH i IN 1..10 %]...[% END %]

[% WHILE count < 10 %]
  [% count = count + 1 %]
[% END %]
```

The `loop` variable provides: `index`, `count`, `first`, `last`, `size`, `max`, `prev`, `next`.

Loop control: `[% NEXT %]` and `[% LAST %]`.

### Template Composition

```
[% INCLUDE header.tt title='Page' %]   # localized variables
[% PROCESS footer.tt %]                # shared variables
[% INSERT rawfile.txt %]               # insert without processing
[% WRAPPER layout.tt %]                # wrap content in a template
  page content here
[% END %]
```

### Blocks and Macros

```
[% BLOCK sidebar %]
  <div class="sidebar">...</div>
[% END %]

[% INCLUDE sidebar %]

[% MACRO greet(name) BLOCK %]
  Hello [% name %]!
[% END %]

[% greet('World') %]
```

### Filters

Inline syntax:

```
[% name | upper %]
[% title | html %]
[% text | truncate(80) %]
[% 'hello' | repeat(3) %]
[% path | uri %]
[% value | trim | upper %]          # chained filters
```

Block syntax:

```
[% FILTER upper %]
  this becomes uppercase
[% END %]
```

### Exception Handling

```
[% TRY %]
  [% THROW 'db' 'connection failed' %]
[% CATCH db %]
  Database error: [% error.info %]
[% CATCH %]
  General error
[% FINAL %]
  Always runs
[% END %]
```

### Expressions

```
[% x + y %]                    # arithmetic: + - * / % div mod
[% 'hello' _ ' ' _ 'world' %] # string concatenation
[% x == y %]                   # comparison: == != < > <= >=
[% a and b %]                  # logical: and or not (also && || !)
[% x ? 'yes' : 'no' %]        # ternary
[% "Hello $name" %]            # string interpolation (double quotes)
[% "path: ${user.name}" %]     # interpolation with braces
```

### Comments

```
[%# this is a comment %]
```

### Whitespace Control

```
[%- trimmed -%]    # trims surrounding whitespace
```

## Built-in Filters

| Filter | Description |
|--------|-------------|
| `html` | Escape HTML entities (`<`, `>`, `&`, `"`) |
| `xml` | Escape XML entities (includes `'`) |
| `uri` / `url` | URI-encode |
| `upper` / `lower` | Case conversion |
| `ucfirst` / `lcfirst` | First character case |
| `trim` | Strip leading/trailing whitespace |
| `collapse` | Collapse internal whitespace to single spaces |
| `indent(n)` | Indent each line by n spaces |
| `truncate(n)` | Truncate to n characters with `...` |
| `repeat(n)` | Repeat n times |
| `replace(old, new)` | String replacement |
| `remove(pattern)` | Remove regex matches |
| `format(fmt)` | Apply format string per line |
| `html_para` | Wrap paragraphs in `<p>` tags |
| `html_break` | Convert double newlines to `<br>` |
| `html_line_break` | Convert all newlines to `<br>` |
| `null` | Discard output |
| `stderr` | Write to stderr |

## Virtual Methods

### Scalar

`length`, `upper`, `lower`, `ucfirst`, `lcfirst`, `trim`, `collapse`, `defined`, `split(sep)`, `replace(old, new)`, `match(pattern)`, `repeat(n)`, `substr(offset, length)`, `chunk(size)`, `dquote`, `list`, `hash`

### List

`size`, `max`, `first`, `last`, `join(sep)`, `sort`, `nsort`, `reverse`, `unique`, `grep(pattern)`, `push(val)`, `pop`, `shift`, `unshift(val)`, `slice(from, to)`, `splice(off, len)`, `merge(list)`, `hash`, `defined(idx)`, `item(idx)`

### Hash

`keys`, `values`, `pairs`, `each`, `list`, `size`, `exists(key)`, `defined(key)`, `delete(key)`, `item(key)`, `import(hash)`, `sort`, `nsort`

## Custom Filters

```go
engine.AddFilter("reverse", func(input string, args []interface{}) string {
    runes := []rune(input)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
})
```

## Custom Virtual Methods

```go
engine.AddVMethod(tt.VMethodScalar, "wordcount", func(val interface{}, args []interface{}) interface{} {
    words := strings.Fields(fmt.Sprint(val))
    return len(words)
})
```

## Configuration

```go
engine := tt.New(tt.Config{
    IncludePath:  []string{"./templates"},
    StartTag:     "[%",       // default
    EndTag:       "%]",       // default
    CacheEnabled: true,       // cache compiled templates
    Absolute:     false,      // allow absolute paths in INCLUDE
    Relative:     false,      // allow relative paths (../) in INCLUDE
})
```

## Go Struct Support

Template variables can be Go structs. Fields and methods are accessible via dot notation:

```go
type User struct {
    Name string
    Age  int
}

func (u User) Greeting() string {
    return "Hello, " + u.Name
}

vars := map[string]interface{}{
    "user": User{Name: "Alice", Age: 30},
}

// In template:
// [% user.Name %] — accesses field
// [% user.Greeting %] — calls method
```

## API Reference

```go
// Create engine
func New(configs ...Config) *Engine

// Process a named template file (requires IncludePath)
func (e *Engine) Process(name string, vars map[string]interface{}, w io.Writer) error

// Process a template string
func (e *Engine) ProcessString(tmpl string, vars map[string]interface{}) (string, error)

// Register a custom filter
func (e *Engine) AddFilter(name string, fn FilterFunc)

// Register a custom virtual method
func (e *Engine) AddVMethod(typ VMethodType, name string, fn VMethodFunc)

// Set a custom template loader
func (e *Engine) SetLoader(loader TemplateLoader)
```

## Testing

```bash
go test ./... -v
```

## License

MIT
