package tt

import (
	"fmt"
	"reflect"
	"strings"
)

// Stash holds template variables with scope support.
type Stash struct {
	vars   map[string]interface{}
	parent *Stash
}

func NewStash(vars map[string]interface{}) *Stash {
	if vars == nil {
		vars = make(map[string]interface{})
	}
	return &Stash{vars: vars}
}

func (s *Stash) Clone() *Stash {
	child := &Stash{
		vars:   make(map[string]interface{}),
		parent: s,
	}
	return child
}

func (s *Stash) Set(key string, val interface{}) {
	s.vars[key] = val
}

func (s *Stash) Get(key string) (interface{}, bool) {
	if val, ok := s.vars[key]; ok {
		return val, true
	}
	if s.parent != nil {
		return s.parent.Get(key)
	}
	return nil, false
}

func (s *Stash) SetDefault(key string, val interface{}) {
	existing, ok := s.Get(key)
	if !ok || !isTruthy(existing) {
		s.vars[key] = val
	}
}

// Resolve resolves a dotted variable path like "foo.bar.baz" with optional args per segment.
func (s *Stash) Resolve(segments []IdentSegment, evalArgs func([]Expr) ([]interface{}, error)) (interface{}, error) {
	if len(segments) == 0 {
		return nil, nil
	}

	name := segments[0].Name
	var args []interface{}
	if evalArgs != nil && len(segments[0].Args) > 0 {
		var err error
		args, err = evalArgs(segments[0].Args)
		if err != nil {
			return nil, err
		}
	}

	val, ok := s.Get(name)
	if !ok {
		val = nil
	}

	if len(args) > 0 && val == nil {
		return nil, nil
	}

	for i := 1; i < len(segments); i++ {
		seg := segments[i]
		var segArgs []interface{}
		if evalArgs != nil && len(seg.Args) > 0 {
			var err error
			segArgs, err = evalArgs(seg.Args)
			if err != nil {
				return nil, err
			}
		}

		key := seg.Name
		if seg.Dynamic {
			// $var — resolve the variable name to get the actual key
			dynVal, _ := s.Get(seg.Name)
			if dynVal != nil {
				key = toString(dynVal)
			}
		}

		var err error
		val, err = dotAccess(val, key, segArgs)
		if err != nil {
			return nil, err
		}
	}

	return val, nil
}

// dotAccess accesses a field/key/method on a value.
func dotAccess(val interface{}, key string, args []interface{}) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(val)

	// For maps, check the key first before vmethods so that user data
	// takes precedence over built-in methods like "size", "keys", etc.
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		result := rv.MapIndex(reflect.ValueOf(key))
		if result.IsValid() {
			return result.Interface(), nil
		}
		// Key not found — fall through to vmethods (size, keys, etc.)
	}

	// Try virtual methods
	if result, handled := tryVMethod(val, key, args); handled {
		return result, nil
	}

	switch rv.Kind() {
	case reflect.Map:
		// Already checked above; key was not found
		return nil, nil

	case reflect.Struct:
		return structAccess(rv, key, args)

	case reflect.Ptr:
		if rv.IsNil() {
			return nil, nil
		}
		elem := rv.Elem()
		if elem.Kind() == reflect.Struct {
			return structAccess(elem, key, args)
		}
		if elem.Kind() == reflect.Map && elem.Type().Key().Kind() == reflect.String {
			result := elem.MapIndex(reflect.ValueOf(key))
			if result.IsValid() {
				return result.Interface(), nil
			}
		}

	case reflect.Slice, reflect.Array:
		if result, handled := tryVMethod(val, key, args); handled {
			return result, nil
		}
	}

	// Try calling method on the value
	method := rv.MethodByName(titleCase(key))
	if method.IsValid() {
		return callMethod(method, args)
	}
	method = rv.MethodByName(key)
	if method.IsValid() {
		return callMethod(method, args)
	}

	return nil, nil
}

func structAccess(rv reflect.Value, key string, args []interface{}) (interface{}, error) {
	// Try exported field
	field := rv.FieldByName(key)
	if field.IsValid() && field.CanInterface() {
		return field.Interface(), nil
	}

	// Try title-cased field name
	titleKey := titleCase(key)
	field = rv.FieldByName(titleKey)
	if field.IsValid() && field.CanInterface() {
		return field.Interface(), nil
	}

	// Case-insensitive field search
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		if strings.EqualFold(t.Field(i).Name, key) {
			f := rv.Field(i)
			if f.CanInterface() {
				return f.Interface(), nil
			}
		}
	}

	// Try methods
	ptrVal := rv.Addr()
	method := ptrVal.MethodByName(titleCase(key))
	if method.IsValid() {
		return callMethod(method, args)
	}
	method = ptrVal.MethodByName(key)
	if method.IsValid() {
		return callMethod(method, args)
	}

	return nil, nil
}

func callMethod(method reflect.Value, args []interface{}) (interface{}, error) {
	mt := method.Type()
	in := make([]reflect.Value, mt.NumIn())

	for i := 0; i < mt.NumIn(); i++ {
		if i < len(args) {
			argVal := reflect.ValueOf(args[i])
			paramType := mt.In(i)
			if argVal.Type().ConvertibleTo(paramType) {
				in[i] = argVal.Convert(paramType)
			} else {
				in[i] = argVal
			}
		} else {
			in[i] = reflect.Zero(mt.In(i))
		}
	}

	results := method.Call(in)
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		return results[0].Interface(), nil
	case 2:
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	default:
		return results[0].Interface(), nil
	}
}

func isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		return rv.String() != "" && rv.String() != "0"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !rv.IsNil()
	default:
		return true
	}
}

func toFloat(val interface{}) (float64, error) {
	if val == nil {
		return 0, nil
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.String:
		s := rv.String()
		var f float64
		_, err := fmt.Sscanf(s, "%f", &f)
		if err != nil {
			return 0, nil
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", val)
	}
}

func toString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func toSlice(val interface{}) []interface{} {
	if val == nil {
		return nil
	}
	if s, ok := val.([]interface{}); ok {
		return s
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		result := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result
	}
	return []interface{}{val}
}

func toMap(val interface{}) map[string]interface{} {
	if val == nil {
		return nil
	}
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		result := make(map[string]interface{}, rv.Len())
		for _, k := range rv.MapKeys() {
			result[k.String()] = rv.MapIndex(k).Interface()
		}
		return result
	}
	return nil
}
