package json

import (
	"io"
	"unicode/utf8"
)

// EncodeOptions controls structural output. MaxDepth zero selects 32.
type EncodeOptions struct {
	Pretty   bool
	Indent   string
	MaxDepth int
}

// Encode writes one JSON value without buffering the complete output.
func Encode(writer io.Writer, value Value, options EncodeOptions) error {
	if options.MaxDepth < 0 {
		return ErrLimits
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 32
	}
	if options.Pretty && options.Indent == "" {
		options.Indent = "  "
	}
	e := encoder{writer: writer, options: options}
	return e.value(value, 0)
}

type encoder struct {
	writer  io.Writer
	options EncodeOptions
}

func (e *encoder) write(text string) error {
	count, err := io.WriteString(e.writer, text)
	if err == nil && count != len(text) {
		return io.ErrShortWrite
	}
	return err
}
func (e *encoder) newline(depth int) error {
	if !e.options.Pretty {
		return nil
	}
	if err := e.write("\n"); err != nil {
		return err
	}
	for range depth {
		if err := e.write(e.options.Indent); err != nil {
			return err
		}
	}
	return nil
}
func (e *encoder) value(value Value, depth int) error {
	switch value.Type {
	case Null:
		return e.write("null")
	case Bool:
		if value.Boolean {
			return e.write("true")
		}
		return e.write("false")
	case Number:
		if !validNumber(value.Text) {
			return SyntaxError{Message: "invalid number value"}
		}
		return e.write(value.Text)
	case String:
		return e.string(value.Text)
	case Array:
		if depth >= e.options.MaxDepth {
			return SyntaxError{Message: "depth limit exceeded"}
		}
		if err := e.write("["); err != nil {
			return err
		}
		for i, item := range value.Elements {
			if i > 0 {
				if err := e.write(","); err != nil {
					return err
				}
			}
			if err := e.newline(depth + 1); err != nil {
				return err
			}
			if err := e.value(item, depth+1); err != nil {
				return err
			}
		}
		if len(value.Elements) > 0 {
			if err := e.newline(depth); err != nil {
				return err
			}
		}
		return e.write("]")
	case Object:
		if depth >= e.options.MaxDepth {
			return SyntaxError{Message: "depth limit exceeded"}
		}
		if err := e.write("{"); err != nil {
			return err
		}
		for i, member := range value.Members {
			if i > 0 {
				if err := e.write(","); err != nil {
					return err
				}
			}
			if err := e.newline(depth + 1); err != nil {
				return err
			}
			if err := e.string(member.Name); err != nil {
				return err
			}
			if e.options.Pretty {
				if err := e.write(": "); err != nil {
					return err
				}
			} else if err := e.write(":"); err != nil {
				return err
			}
			if err := e.value(member.Value, depth+1); err != nil {
				return err
			}
		}
		if len(value.Members) > 0 {
			if err := e.newline(depth); err != nil {
				return err
			}
		}
		return e.write("}")
	default:
		return SyntaxError{Message: "invalid value kind"}
	}
}
func (e *encoder) string(value string) error {
	if !utf8.ValidString(value) {
		return SyntaxError{Message: "invalid UTF-8 string"}
	}
	if err := e.write("\""); err != nil {
		return err
	}
	start := 0
	for i := 0; i < len(value); i++ {
		b := value[i]
		var escape string
		switch b {
		case '"':
			escape = "\\\""
		case '\\':
			escape = "\\\\"
		case '\b':
			escape = "\\b"
		case '\f':
			escape = "\\f"
		case '\n':
			escape = "\\n"
		case '\r':
			escape = "\\r"
		case '\t':
			escape = "\\t"
		default:
			if b < 0x20 {
				const hex = "0123456789abcdef"
				escape = "\\u00" + string([]byte{hex[b>>4], hex[b&15]})
			}
		}
		if escape != "" {
			if start < i {
				if err := e.write(value[start:i]); err != nil {
					return err
				}
			}
			if err := e.write(escape); err != nil {
				return err
			}
			start = i + 1
		}
	}
	if start < len(value) {
		if err := e.write(value[start:]); err != nil {
			return err
		}
	}
	return e.write("\"")
}
func validNumber(value string) bool {
	p := parser{data: []byte(value), limits: Limits{MaxStringBytes: 1}}
	_, err := p.number()
	return err == nil && p.pos == len(value)
}
