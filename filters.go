package tt

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// FilterFunc is the signature for filter functions.
// A filter takes input text and optional args, returns filtered text.
type FilterFunc func(input string, args []interface{}) string

var defaultFilters = map[string]FilterFunc{}

func init() {
	registerDefaultFilters()
}

func registerDefaultFilters() {
	defaultFilters["html"] = filterHTML
	defaultFilters["html_entity"] = filterHTML
	defaultFilters["xml"] = filterXML
	defaultFilters["html_para"] = filterHTMLPara
	defaultFilters["html_break"] = filterHTMLBreak
	defaultFilters["html_line_break"] = filterHTMLLineBreak
	defaultFilters["uri"] = filterURI
	defaultFilters["url"] = filterURL
	defaultFilters["upper"] = filterUpper
	defaultFilters["lower"] = filterLower
	defaultFilters["ucfirst"] = filterUcfirst
	defaultFilters["lcfirst"] = filterLcfirst
	defaultFilters["trim"] = filterTrim
	defaultFilters["collapse"] = filterCollapse
	defaultFilters["indent"] = filterIndent
	defaultFilters["truncate"] = filterTruncate
	defaultFilters["repeat"] = filterRepeat
	defaultFilters["replace"] = filterReplace
	defaultFilters["remove"] = filterRemove
	defaultFilters["format"] = filterFormat
	defaultFilters["null"] = filterNull
	defaultFilters["stderr"] = filterStderr
}

func filterHTML(input string, args []interface{}) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(input)
}

func filterXML(input string, args []interface{}) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(input)
}

func filterHTMLPara(input string, args []interface{}) string {
	re := regexp.MustCompile(`\n{2,}`)
	paragraphs := re.Split(strings.TrimSpace(input), -1)
	var result []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, "<p>\n"+p+"\n</p>")
		}
	}
	return strings.Join(result, "\n\n")
}

func filterHTMLBreak(input string, args []interface{}) string {
	re := regexp.MustCompile(`\n{2,}`)
	return re.ReplaceAllString(input, "<br>\n<br>\n")
}

func filterHTMLLineBreak(input string, args []interface{}) string {
	return strings.ReplaceAll(input, "\n", "<br>\n")
}

func filterURI(input string, args []interface{}) string {
	return url.QueryEscape(input)
}

func filterURL(input string, args []interface{}) string {
	return url.QueryEscape(input)
}

func filterUpper(input string, args []interface{}) string {
	return strings.ToUpper(input)
}

func filterLower(input string, args []interface{}) string {
	return strings.ToLower(input)
}

func filterUcfirst(input string, args []interface{}) string {
	if len(input) == 0 {
		return input
	}
	return strings.ToUpper(input[:1]) + input[1:]
}

func filterLcfirst(input string, args []interface{}) string {
	if len(input) == 0 {
		return input
	}
	return strings.ToLower(input[:1]) + input[1:]
}

func filterTrim(input string, args []interface{}) string {
	return strings.TrimSpace(input)
}

func filterCollapse(input string, args []interface{}) string {
	s := strings.TrimSpace(input)
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

func filterIndent(input string, args []interface{}) string {
	n := 4
	pad := " "
	if len(args) > 0 {
		if f, err := toFloat(args[0]); err == nil {
			n = int(f)
		} else {
			pad = toString(args[0])
			n = 1
		}
	}
	if len(args) > 1 {
		pad = toString(args[1])
	}
	prefix := strings.Repeat(pad, n)
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

func filterTruncate(input string, args []interface{}) string {
	length := 32
	if len(args) > 0 {
		if f, err := toFloat(args[0]); err == nil {
			length = int(f)
		}
	}
	suffix := "..."
	if len(args) > 1 {
		suffix = toString(args[1])
	}
	if len(input) <= length {
		return input
	}
	return input[:length] + suffix
}

func filterRepeat(input string, args []interface{}) string {
	n := 1
	if len(args) > 0 {
		if f, err := toFloat(args[0]); err == nil {
			n = int(f)
		}
	}
	return strings.Repeat(input, n)
}

func filterReplace(input string, args []interface{}) string {
	if len(args) < 1 {
		return input
	}
	old := toString(args[0])
	newStr := ""
	if len(args) > 1 {
		newStr = toString(args[1])
	}
	return strings.ReplaceAll(input, old, newStr)
}

func filterRemove(input string, args []interface{}) string {
	if len(args) < 1 {
		return input
	}
	pattern := toString(args[0])
	re, err := regexp.Compile(pattern)
	if err != nil {
		return input
	}
	return re.ReplaceAllString(input, "")
}

func filterFormat(input string, args []interface{}) string {
	if len(args) < 1 {
		return input
	}
	format := toString(args[0])
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = fmt.Sprintf(format, line)
	}
	return strings.Join(lines, "\n")
}

func filterNull(input string, args []interface{}) string {
	return ""
}

func filterStderr(input string, args []interface{}) string {
	fmt.Fprint(os.Stderr, input)
	return ""
}
