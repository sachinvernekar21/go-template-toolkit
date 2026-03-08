package tt

import "fmt"

type TemplateError struct {
	Type    string
	Message string
	Line    int
}

func (e *TemplateError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s error (line %d): %s", e.Type, e.Line, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

func newParseError(line int, msg string) *TemplateError {
	return &TemplateError{Type: "parse", Message: msg, Line: line}
}

func newEvalError(msg string) *TemplateError {
	return &TemplateError{Type: "eval", Message: msg}
}

func newFileError(msg string) *TemplateError {
	return &TemplateError{Type: "file", Message: msg}
}

// FlowSignal is used for loop control and early returns — not a real error.
type FlowSignal struct {
	Action string // "NEXT", "LAST", "RETURN", "STOP"
}

func (f *FlowSignal) Error() string {
	return f.Action
}

// ThrowSignal carries a TRY/CATCH exception.
type ThrowSignal struct {
	Type    string
	Message string
}

func (t *ThrowSignal) Error() string {
	if t.Type != "" {
		return fmt.Sprintf("%s: %s", t.Type, t.Message)
	}
	return t.Message
}
