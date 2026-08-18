// Package json provides bounded JSON decoding and encoding for Playdate games.
// It avoids reflection, callbacks, and deferred cleanup so the same codec can
// run in host tests, the Simulator, and the conservative device profile.
package json

import (
	"errors"
	"io"
	"unicode/utf8"
)

// Kind identifies one JSON value kind.
type Kind uint8

const (
	Invalid Kind = iota
	Null
	Bool
	Number
	String
	Array
	Object
)

// Member preserves one object member. Object order and duplicate names are
// retained exactly as decoded.
type Member struct {
	Name  string
	Value Value
}

// Value is a reflection-free JSON tree. Text contains decoded string data or
// the original JSON number spelling, depending on Type.
type Value struct {
	Type     Kind
	Boolean  bool
	Text     string
	Elements []Value
	Members  []Member
}

// Lookup returns the first member with name.
func (v Value) Lookup(name string) (Value, bool) {
	if v.Type != Object {
		return Value{}, false
	}
	for _, member := range v.Members {
		if member.Name == name {
			return member.Value, true
		}
	}
	return Value{}, false
}

// Limits bounds decoder memory and nesting. Zero fields select conservative
// defaults. MaxBytes includes the complete encoded document.
type Limits struct {
	MaxBytes       int
	MaxDepth       int
	MaxNodes       int
	MaxStringBytes int
}

var ErrLimits = errors.New("invalid JSON limits")

// SyntaxError reports a byte offset, JSON Pointer path, and stable diagnostic.
type SyntaxError struct {
	Offset        int
	Path, Message string
}

func (e SyntaxError) Error() string {
	if e.Path == "" {
		return "json at byte " + decimal(e.Offset) + ": " + e.Message
	}
	return "json at byte " + decimal(e.Offset) + " (" + e.Path + "): " + e.Message
}

func decimal(value int) string {
	if value == 0 {
		return "0"
	}
	var buf [24]byte
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[index:])
}

func normalized(l Limits) (Limits, error) {
	if l.MaxBytes < 0 || l.MaxDepth < 0 || l.MaxNodes < 0 || l.MaxStringBytes < 0 {
		return Limits{}, ErrLimits
	}
	if l.MaxBytes == 0 {
		l.MaxBytes = 64 * 1024
	}
	if l.MaxDepth == 0 {
		l.MaxDepth = 32
	}
	if l.MaxNodes == 0 {
		l.MaxNodes = 4096
	}
	if l.MaxStringBytes == 0 {
		l.MaxStringBytes = 16 * 1024
	}
	return l, nil
}

// Decode reads and parses one complete document under limits.
func Decode(reader io.Reader, limits Limits) (Value, error) {
	limits, err := normalized(limits)
	if err != nil {
		return Value{}, err
	}
	data := make([]byte, 0, min(limits.MaxBytes, 1024))
	var chunk [512]byte
	empty := 0
	for {
		count, readErr := reader.Read(chunk[:])
		if count > 0 {
			empty = 0
			if len(data)+count > limits.MaxBytes {
				return Value{}, SyntaxError{Offset: len(data), Message: "document exceeds byte limit"}
			}
			data = append(data, chunk[:count]...)
		} else if readErr == nil {
			empty++
			if empty >= 100 {
				return Value{}, io.ErrNoProgress
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				return Value{}, readErr
			}
			break
		}
	}
	return DecodeBytes(data, limits)
}

// DecodeBytes parses one complete document. The returned tree owns its strings.
func DecodeBytes(data []byte, limits Limits) (Value, error) {
	limits, err := normalized(limits)
	if err != nil {
		return Value{}, err
	}
	if len(data) > limits.MaxBytes {
		return Value{}, SyntaxError{Offset: limits.MaxBytes, Message: "document exceeds byte limit"}
	}
	p := parser{data: data, limits: limits}
	p.space()
	value, err := p.value(0)
	if err != nil {
		return Value{}, err
	}
	p.space()
	if p.pos != len(data) {
		return Value{}, p.failure("unexpected trailing data")
	}
	return value, nil
}

type parser struct {
	data       []byte
	pos, nodes int
	limits     Limits
	path       []string
}

func (p *parser) failure(message string) error {
	return SyntaxError{Offset: p.pos, Path: pointer(p.path), Message: message}
}
func pointer(parts []string) string {
	var result string
	for _, part := range parts {
		result += "/"
		for i := 0; i < len(part); i++ {
			if part[i] == '~' {
				result += "~0"
			} else if part[i] == '/' {
				result += "~1"
			} else {
				result += string(part[i])
			}
		}
	}
	return result
}
func (p *parser) space() {
	for p.pos < len(p.data) {
		switch p.data[p.pos] {
		case ' ', '\t', '\r', '\n':
			p.pos++
		default:
			return
		}
	}
}
func (p *parser) value(depth int) (Value, error) {
	if p.nodes >= p.limits.MaxNodes {
		return Value{}, p.failure("node limit exceeded")
	}
	p.nodes++
	if p.pos >= len(p.data) {
		return Value{}, p.failure("expected value")
	}
	switch p.data[p.pos] {
	case 'n':
		if !p.literal("null") {
			return Value{}, p.failure("invalid literal")
		}
		return Value{Type: Null}, nil
	case 't':
		if !p.literal("true") {
			return Value{}, p.failure("invalid literal")
		}
		return Value{Type: Bool, Boolean: true}, nil
	case 'f':
		if !p.literal("false") {
			return Value{}, p.failure("invalid literal")
		}
		return Value{Type: Bool}, nil
	case '"':
		text, err := p.string()
		return Value{Type: String, Text: text}, err
	case '[':
		return p.array(depth + 1)
	case '{':
		return p.object(depth + 1)
	default:
		if p.data[p.pos] == '-' || (p.data[p.pos] >= '0' && p.data[p.pos] <= '9') {
			text, err := p.number()
			return Value{Type: Number, Text: text}, err
		}
		return Value{}, p.failure("expected value")
	}
}
func (p *parser) literal(value string) bool {
	if len(p.data)-p.pos < len(value) || string(p.data[p.pos:p.pos+len(value)]) != value {
		return false
	}
	p.pos += len(value)
	return true
}
func (p *parser) array(depth int) (Value, error) {
	if depth > p.limits.MaxDepth {
		return Value{}, p.failure("depth limit exceeded")
	}
	p.pos++
	p.space()
	result := Value{Type: Array}
	if p.take(']') {
		return result, nil
	}
	index := 0
	for {
		p.path = append(p.path, decimal(index))
		v, err := p.value(depth)
		p.path = p.path[:len(p.path)-1]
		if err != nil {
			return Value{}, err
		}
		result.Elements = append(result.Elements, v)
		p.space()
		if p.take(']') {
			return result, nil
		}
		if !p.take(',') {
			return Value{}, p.failure("expected comma or array end")
		}
		p.space()
		index++
	}
}
func (p *parser) object(depth int) (Value, error) {
	if depth > p.limits.MaxDepth {
		return Value{}, p.failure("depth limit exceeded")
	}
	p.pos++
	p.space()
	result := Value{Type: Object}
	if p.take('}') {
		return result, nil
	}
	for {
		if p.pos >= len(p.data) || p.data[p.pos] != '"' {
			return Value{}, p.failure("expected object member name")
		}
		name, err := p.string()
		if err != nil {
			return Value{}, err
		}
		p.space()
		if !p.take(':') {
			return Value{}, p.failure("expected colon")
		}
		p.space()
		p.path = append(p.path, name)
		v, err := p.value(depth)
		p.path = p.path[:len(p.path)-1]
		if err != nil {
			return Value{}, err
		}
		result.Members = append(result.Members, Member{Name: name, Value: v})
		p.space()
		if p.take('}') {
			return result, nil
		}
		if !p.take(',') {
			return Value{}, p.failure("expected comma or object end")
		}
		p.space()
	}
}
func (p *parser) take(value byte) bool {
	if p.pos < len(p.data) && p.data[p.pos] == value {
		p.pos++
		return true
	}
	return false
}
func (p *parser) string() (string, error) {
	p.pos++
	out := make([]byte, 0, 32)
	for p.pos < len(p.data) {
		b := p.data[p.pos]
		p.pos++
		if b == '"' {
			if len(out) > p.limits.MaxStringBytes {
				return "", p.failure("string exceeds byte limit")
			}
			return string(out), nil
		}
		if b < 0x20 {
			return "", p.failure("control character in string")
		}
		if b == '\\' {
			if p.pos >= len(p.data) {
				return "", p.failure("incomplete escape")
			}
			e := p.data[p.pos]
			p.pos++
			switch e {
			case '"', '\\', '/':
				out = append(out, e)
			case 'b':
				out = append(out, '\b')
			case 'f':
				out = append(out, '\f')
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case 'u':
				r, err := p.unicode()
				if err != nil {
					return "", err
				}
				out = utf8.AppendRune(out, r)
			default:
				return "", p.failure("invalid escape")
			}
		} else {
			if b < utf8.RuneSelf {
				out = append(out, b)
			} else {
				p.pos--
				r, size := utf8.DecodeRune(p.data[p.pos:])
				if r == utf8.RuneError && size == 1 {
					return "", p.failure("invalid UTF-8")
				}
				out = append(out, p.data[p.pos:p.pos+size]...)
				p.pos += size
			}
		}
		if len(out) > p.limits.MaxStringBytes {
			return "", p.failure("string exceeds byte limit")
		}
	}
	return "", p.failure("unterminated string")
}
func (p *parser) unicode() (rune, error) {
	first, ok := p.hex4()
	if !ok {
		return 0, p.failure("invalid Unicode escape")
	}
	if first >= 0xD800 && first <= 0xDBFF {
		if p.pos+2 > len(p.data) || p.data[p.pos] != '\\' || p.data[p.pos+1] != 'u' {
			return 0, p.failure("missing low surrogate")
		}
		p.pos += 2
		second, ok := p.hex4()
		if !ok || second < 0xDC00 || second > 0xDFFF {
			return 0, p.failure("invalid low surrogate")
		}
		return rune(0x10000 + (first-0xD800)*0x400 + (second - 0xDC00)), nil
	}
	if first >= 0xDC00 && first <= 0xDFFF {
		return 0, p.failure("unexpected low surrogate")
	}
	return rune(first), nil
}
func (p *parser) hex4() (int, bool) {
	if p.pos+4 > len(p.data) {
		return 0, false
	}
	value := 0
	for range 4 {
		b := p.data[p.pos]
		p.pos++
		value *= 16
		switch {
		case b >= '0' && b <= '9':
			value += int(b - '0')
		case b >= 'a' && b <= 'f':
			value += int(b-'a') + 10
		case b >= 'A' && b <= 'F':
			value += int(b-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
func (p *parser) number() (string, error) {
	start := p.pos
	if p.take('-') && p.pos >= len(p.data) {
		return "", p.failure("incomplete number")
	}
	if p.take('0') {
		if p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '9' {
			return "", p.failure("leading zero in number")
		}
	} else {
		if p.pos >= len(p.data) || p.data[p.pos] < '1' || p.data[p.pos] > '9' {
			return "", p.failure("invalid number")
		}
		for p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '9' {
			p.pos++
		}
	}
	if p.take('.') {
		if p.pos >= len(p.data) || p.data[p.pos] < '0' || p.data[p.pos] > '9' {
			return "", p.failure("missing fraction digits")
		}
		for p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '9' {
			p.pos++
		}
	}
	if p.pos < len(p.data) && (p.data[p.pos] == 'e' || p.data[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.data) && (p.data[p.pos] == '+' || p.data[p.pos] == '-') {
			p.pos++
		}
		if p.pos >= len(p.data) || p.data[p.pos] < '0' || p.data[p.pos] > '9' {
			return "", p.failure("missing exponent digits")
		}
		for p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '9' {
			p.pos++
		}
	}
	return string(p.data[start:p.pos]), nil
}
