package tt

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// VMethodType identifies scalar, list, or hash virtual methods.
type VMethodType int

const (
	VMethodScalar VMethodType = iota
	VMethodList
	VMethodHash
)

// VMethodFunc is the signature for user-registered vmethods.
type VMethodFunc func(val interface{}, args []interface{}) interface{}

var (
	scalarVMethods = map[string]VMethodFunc{}
	listVMethods   = map[string]VMethodFunc{}
	hashVMethods   = map[string]VMethodFunc{}
)

func init() {
	registerScalarVMethods()
	registerListVMethods()
	registerHashVMethods()
}

// RegisterVMethod registers a custom virtual method.
func RegisterVMethod(typ VMethodType, name string, fn VMethodFunc) {
	switch typ {
	case VMethodScalar:
		scalarVMethods[name] = fn
	case VMethodList:
		listVMethods[name] = fn
	case VMethodHash:
		hashVMethods[name] = fn
	}
}

// tryVMethod attempts to call a virtual method on a value. Returns (result, true) if handled.
func tryVMethod(val interface{}, name string, args []interface{}) (interface{}, bool) {
	rv := reflect.ValueOf(val)

	switch {
	case rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array:
		if fn, ok := listVMethods[name]; ok {
			return fn(val, args), true
		}
	case rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String:
		if fn, ok := hashVMethods[name]; ok {
			return fn(val, args), true
		}
	}

	if fn, ok := scalarVMethods[name]; ok {
		return fn(val, args), true
	}

	return nil, false
}

func registerScalarVMethods() {
	scalarVMethods["length"] = func(val interface{}, args []interface{}) interface{} {
		return len(toString(val))
	}

	scalarVMethods["upper"] = func(val interface{}, args []interface{}) interface{} {
		return strings.ToUpper(toString(val))
	}

	scalarVMethods["lower"] = func(val interface{}, args []interface{}) interface{} {
		return strings.ToLower(toString(val))
	}

	scalarVMethods["ucfirst"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}

	scalarVMethods["lcfirst"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		if len(s) == 0 {
			return s
		}
		return strings.ToLower(s[:1]) + s[1:]
	}

	scalarVMethods["trim"] = func(val interface{}, args []interface{}) interface{} {
		return strings.TrimSpace(toString(val))
	}

	scalarVMethods["collapse"] = func(val interface{}, args []interface{}) interface{} {
		s := strings.TrimSpace(toString(val))
		re := regexp.MustCompile(`\s+`)
		return re.ReplaceAllString(s, " ")
	}

	scalarVMethods["defined"] = func(val interface{}, args []interface{}) interface{} {
		return val != nil
	}

	scalarVMethods["split"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		sep := ""
		if len(args) > 0 {
			sep = toString(args[0])
		}
		if sep == "" {
			parts := make([]interface{}, len(s))
			for i, ch := range s {
				parts[i] = string(ch)
			}
			return parts
		}
		parts := strings.Split(s, sep)
		result := make([]interface{}, len(parts))
		for i, p := range parts {
			result[i] = p
		}
		return result
	}

	scalarVMethods["replace"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		if len(args) < 1 {
			return s
		}
		old := toString(args[0])
		newStr := ""
		if len(args) > 1 {
			newStr = toString(args[1])
		}
		return strings.ReplaceAll(s, old, newStr)
	}

	scalarVMethods["match"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		if len(args) < 1 {
			return nil
		}
		pattern := toString(args[0])
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil
		}
		matches := re.FindStringSubmatch(s)
		if matches == nil {
			return nil
		}
		if len(matches) > 1 {
			result := make([]interface{}, len(matches)-1)
			for i, m := range matches[1:] {
				result[i] = m
			}
			return result
		}
		return true
	}

	scalarVMethods["repeat"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		n := 1
		if len(args) > 0 {
			if f, err := toFloat(args[0]); err == nil {
				n = int(f)
			}
		}
		return strings.Repeat(s, n)
	}

	scalarVMethods["substr"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		if len(args) < 1 {
			return s
		}
		off := 0
		if f, err := toFloat(args[0]); err == nil {
			off = int(f)
		}
		if off >= len(s) {
			return ""
		}
		if off < 0 {
			off = len(s) + off
			if off < 0 {
				off = 0
			}
		}
		if len(args) >= 2 {
			length := len(s) - off
			if f, err := toFloat(args[1]); err == nil {
				length = int(f)
			}
			end := off + length
			if end > len(s) {
				end = len(s)
			}
			return s[off:end]
		}
		return s[off:]
	}

	scalarVMethods["list"] = func(val interface{}, args []interface{}) interface{} {
		return []interface{}{val}
	}

	scalarVMethods["hash"] = func(val interface{}, args []interface{}) interface{} {
		return map[string]interface{}{"value": val}
	}

	scalarVMethods["chunk"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		size := 1
		if len(args) > 0 {
			if f, err := toFloat(args[0]); err == nil && f > 0 {
				size = int(f)
			}
		}
		var chunks []interface{}
		for i := 0; i < len(s); i += size {
			end := i + size
			if end > len(s) {
				end = len(s)
			}
			chunks = append(chunks, s[i:end])
		}
		return chunks
	}

	scalarVMethods["dquote"] = func(val interface{}, args []interface{}) interface{} {
		s := toString(val)
		s = strings.ReplaceAll(s, `"`, `\"`)
		s = strings.ReplaceAll(s, "\n", `\n`)
		return s
	}
}

func registerListVMethods() {
	listVMethods["size"] = func(val interface{}, args []interface{}) interface{} {
		return len(toSlice(val))
	}

	listVMethods["max"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		if len(s) == 0 {
			return 0
		}
		return len(s) - 1
	}

	listVMethods["first"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		n := 1
		if len(args) > 0 {
			if f, err := toFloat(args[0]); err == nil {
				n = int(f)
			}
		}
		if n == 1 {
			if len(s) > 0 {
				return s[0]
			}
			return nil
		}
		if n > len(s) {
			n = len(s)
		}
		return s[:n]
	}

	listVMethods["last"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		n := 1
		if len(args) > 0 {
			if f, err := toFloat(args[0]); err == nil {
				n = int(f)
			}
		}
		if n == 1 {
			if len(s) > 0 {
				return s[len(s)-1]
			}
			return nil
		}
		if n > len(s) {
			n = len(s)
		}
		return s[len(s)-n:]
	}

	listVMethods["join"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		sep := " "
		if len(args) > 0 {
			sep = toString(args[0])
		}
		parts := make([]string, len(s))
		for i, v := range s {
			parts[i] = toString(v)
		}
		return strings.Join(parts, sep)
	}

	listVMethods["sort"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		result := make([]interface{}, len(s))
		copy(result, s)

		var keyField string
		if len(args) > 0 {
			keyField = toString(args[0])
		}

		sort.Slice(result, func(i, j int) bool {
			a, b := result[i], result[j]
			if keyField != "" {
				a = extractField(a, keyField)
				b = extractField(b, keyField)
			}
			return toString(a) < toString(b)
		})
		return result
	}

	listVMethods["nsort"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		result := make([]interface{}, len(s))
		copy(result, s)

		var keyField string
		if len(args) > 0 {
			keyField = toString(args[0])
		}

		sort.Slice(result, func(i, j int) bool {
			a, b := result[i], result[j]
			if keyField != "" {
				a = extractField(a, keyField)
				b = extractField(b, keyField)
			}
			af, _ := toFloat(a)
			bf, _ := toFloat(b)
			return af < bf
		})
		return result
	}

	listVMethods["reverse"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		result := make([]interface{}, len(s))
		for i, v := range s {
			result[len(s)-1-i] = v
		}
		return result
	}

	listVMethods["unique"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		seen := make(map[string]bool)
		var result []interface{}
		for _, v := range s {
			key := toString(v)
			if !seen[key] {
				seen[key] = true
				result = append(result, v)
			}
		}
		return result
	}

	listVMethods["grep"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		if len(args) < 1 {
			return s
		}
		pattern := toString(args[0])
		re, err := regexp.Compile(pattern)
		if err != nil {
			return s
		}
		var result []interface{}
		for _, v := range s {
			if re.MatchString(toString(v)) {
				result = append(result, v)
			}
		}
		return result
	}

	listVMethods["push"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		return append(s, args...)
	}

	listVMethods["pop"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		if len(s) == 0 {
			return nil
		}
		return s[len(s)-1]
	}

	listVMethods["shift"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		if len(s) == 0 {
			return nil
		}
		return s[0]
	}

	listVMethods["unshift"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		return append(args, s...)
	}

	listVMethods["slice"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		from, to := 0, len(s)-1
		if len(args) > 0 {
			if f, err := toFloat(args[0]); err == nil {
				from = int(f)
			}
		}
		if len(args) > 1 {
			if f, err := toFloat(args[1]); err == nil {
				to = int(f)
			}
		}
		if from < 0 {
			from = 0
		}
		if to >= len(s) {
			to = len(s) - 1
		}
		if from > to {
			return []interface{}{}
		}
		return s[from : to+1]
	}

	listVMethods["defined"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		if len(args) > 0 {
			if f, err := toFloat(args[0]); err == nil {
				idx := int(f)
				return idx >= 0 && idx < len(s) && s[idx] != nil
			}
		}
		return val != nil
	}

	listVMethods["hash"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		result := make(map[string]interface{})
		for i := 0; i+1 < len(s); i += 2 {
			key := toString(s[i])
			result[key] = s[i+1]
		}
		return result
	}

	listVMethods["merge"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		result := make([]interface{}, len(s))
		copy(result, s)
		for _, arg := range args {
			result = append(result, toSlice(arg)...)
		}
		return result
	}

	listVMethods["item"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		if len(args) > 0 {
			if f, err := toFloat(args[0]); err == nil {
				idx := int(f)
				if idx >= 0 && idx < len(s) {
					return s[idx]
				}
			}
		}
		return nil
	}

	listVMethods["splice"] = func(val interface{}, args []interface{}) interface{} {
		s := toSlice(val)
		if len(args) < 1 {
			return s
		}
		off := 0
		if f, err := toFloat(args[0]); err == nil {
			off = int(f)
		}
		length := len(s) - off
		if len(args) > 1 {
			if f, err := toFloat(args[1]); err == nil {
				length = int(f)
			}
		}
		end := off + length
		if end > len(s) {
			end = len(s)
		}
		removed := s[off:end]
		result := make([]interface{}, 0, len(s)-length+len(args)-2)
		result = append(result, s[:off]...)
		if len(args) > 2 {
			result = append(result, args[2:]...)
		}
		result = append(result, s[end:]...)
		_ = result
		return removed
	}
}

func registerHashVMethods() {
	hashVMethods["keys"] = func(val interface{}, args []interface{}) interface{} {
		m := toMap(val)
		keys := make([]interface{}, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}

	hashVMethods["values"] = func(val interface{}, args []interface{}) interface{} {
		m := toMap(val)
		vals := make([]interface{}, 0, len(m))
		for _, v := range m {
			vals = append(vals, v)
		}
		return vals
	}

	hashVMethods["pairs"] = func(val interface{}, args []interface{}) interface{} {
		m := toMap(val)
		pairs := make([]interface{}, 0, len(m))
		keys := sortedKeys(m)
		for _, k := range keys {
			pairs = append(pairs, map[string]interface{}{
				"key":   k,
				"value": m[k],
			})
		}
		return pairs
	}

	hashVMethods["each"] = func(val interface{}, args []interface{}) interface{} {
		m := toMap(val)
		var result []interface{}
		for k, v := range m {
			result = append(result, k, v)
		}
		return result
	}

	hashVMethods["list"] = func(val interface{}, args []interface{}) interface{} {
		return hashVMethods["pairs"](val, args)
	}

	hashVMethods["size"] = func(val interface{}, args []interface{}) interface{} {
		return len(toMap(val))
	}

	hashVMethods["exists"] = func(val interface{}, args []interface{}) interface{} {
		if len(args) < 1 {
			return false
		}
		m := toMap(val)
		_, ok := m[toString(args[0])]
		return ok
	}

	hashVMethods["defined"] = func(val interface{}, args []interface{}) interface{} {
		if len(args) < 1 {
			return val != nil
		}
		m := toMap(val)
		v, ok := m[toString(args[0])]
		return ok && v != nil
	}

	hashVMethods["delete"] = func(val interface{}, args []interface{}) interface{} {
		m := toMap(val)
		for _, arg := range args {
			delete(m, toString(arg))
		}
		return ""
	}

	hashVMethods["item"] = func(val interface{}, args []interface{}) interface{} {
		if len(args) < 1 {
			return nil
		}
		m := toMap(val)
		return m[toString(args[0])]
	}

	hashVMethods["import"] = func(val interface{}, args []interface{}) interface{} {
		m := toMap(val)
		for _, arg := range args {
			other := toMap(arg)
			for k, v := range other {
				m[k] = v
			}
		}
		return ""
	}

	hashVMethods["sort"] = func(val interface{}, args []interface{}) interface{} {
		m := toMap(val)
		keys := sortedKeys(m)
		result := make([]interface{}, len(keys))
		for i, k := range keys {
			result[i] = k
		}
		return result
	}

	hashVMethods["nsort"] = func(val interface{}, args []interface{}) interface{} {
		m := toMap(val)
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			a, _ := toFloat(m[keys[i]])
			b, _ := toFloat(m[keys[j]])
			return a < b
		})
		result := make([]interface{}, len(keys))
		for i, k := range keys {
			result[i] = k
		}
		return result
	}
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func extractField(val interface{}, field string) interface{} {
	if m, ok := val.(map[string]interface{}); ok {
		return m[field]
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Struct {
		f := rv.FieldByName(field)
		if f.IsValid() && f.CanInterface() {
			return f.Interface()
		}
		titleField := titleCase(field)
		f = rv.FieldByName(titleField)
		if f.IsValid() && f.CanInterface() {
			return f.Interface()
		}
	}
	return val
}

// Used in evaluator for range expressions
func makeRange(start, end int) []interface{} {
	if start > end {
		result := make([]interface{}, start-end+1)
		for i := start; i >= end; i-- {
			result[start-i] = i
		}
		return result
	}
	result := make([]interface{}, end-start+1)
	for i := start; i <= end; i++ {
		result[i-start] = i
	}
	return result
}

// Used for integer division
func intDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return math.Floor(a / b)
}

// Used for modulo
func fmod(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return math.Mod(a, b)
}

func toInt(val interface{}) int {
	f, _ := toFloat(val)
	return int(f)
}

func formatNumber(f float64) string {
	if f == math.Trunc(f) && !math.IsInf(f, 0) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%g", f)
}
