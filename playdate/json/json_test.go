package json

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestDecodeEncodeRoundTrip(t *testing.T) {
	input := `{"name":"Crank \\u0026 Key","emoji":"\uD83D\uDE80","enabled":true,"nothing":null,"number":-12.50e+2,"items":[1,"two",false],"same":1,"same":2}`
	value, err := Decode(strings.NewReader(input), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	name, ok := value.Lookup("name")
	if !ok || name.Text != "Crank \\u0026 Key" {
		t.Fatalf("name=%+v", name)
	}
	emoji, _ := value.Lookup("emoji")
	if emoji.Text != "🚀" {
		t.Fatalf("emoji=%q", emoji.Text)
	}
	number, _ := value.Lookup("number")
	if number.Text != "-12.50e+2" {
		t.Fatalf("number=%q", number.Text)
	}
	if len(value.Members) != 8 || value.Members[6].Name != "same" || value.Members[7].Name != "same" {
		t.Fatalf("members=%+v", value.Members)
	}
	var output bytes.Buffer
	if err := Encode(&output, value, EncodeOptions{}); err != nil {
		t.Fatal(err)
	}
	again, err := DecodeBytes(output.Bytes(), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Members) != len(value.Members) {
		t.Fatalf("round trip members=%d", len(again.Members))
	}
}

func TestDecodeSyntaxPathsAndLimits(t *testing.T) {
	tests := []struct{ input, want string }{
		{`{"a":[0,]}`, "/a/1"}, {`{"a":"\uD800x"}`, "/a"}, {`01`, "leading zero"}, {`true false`, "trailing"},
	}
	for _, test := range tests {
		_, err := DecodeBytes([]byte(test.input), Limits{})
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("Decode(%q)=%v, want %q", test.input, err, test.want)
		}
	}
	if _, err := DecodeBytes([]byte(`[[]]`), Limits{MaxDepth: 1}); err == nil || !strings.Contains(err.Error(), "depth") {
		t.Fatalf("depth=%v", err)
	}
	if _, err := DecodeBytes([]byte(`[1,2]`), Limits{MaxNodes: 2}); err == nil || !strings.Contains(err.Error(), "node") {
		t.Fatalf("nodes=%v", err)
	}
	if _, err := DecodeBytes([]byte(`"long"`), Limits{MaxStringBytes: 3}); err == nil || !strings.Contains(err.Error(), "string") {
		t.Fatalf("string=%v", err)
	}
	if _, err := DecodeBytes([]byte(`null`), Limits{MaxBytes: 3}); err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("bytes=%v", err)
	}
	if _, err := DecodeBytes(nil, Limits{MaxDepth: -1}); !errors.Is(err, ErrLimits) {
		t.Fatalf("limits=%v", err)
	}
	if _, err := DecodeBytes([]byte{'"', 0xff, '"'}, Limits{}); err == nil || !strings.Contains(err.Error(), "invalid UTF-8") {
		t.Fatalf("UTF-8=%v", err)
	}
}

type failingReader struct{ done bool }

func (r *failingReader) Read(p []byte) (int, error) {
	if !r.done {
		r.done = true
		copy(p, "nu")
		return 2, nil
	}
	return 0, io.ErrUnexpectedEOF
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

type shortWriter struct{}

func (shortWriter) Write(value []byte) (int, error) {
	if len(value) == 0 {
		return 0, nil
	}
	return len(value) - 1, nil
}

func TestIOFailuresAndPrettyEncoding(t *testing.T) {
	if _, err := Decode(&failingReader{}, Limits{}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("read=%v", err)
	}
	value := Value{Type: Object, Members: []Member{{Name: "line", Value: Value{Type: String, Text: "a\n\tb"}}, {Name: "values", Value: Value{Type: Array, Elements: []Value{{Type: Number, Text: "1"}, {Type: Bool, Boolean: true}}}}}}
	var output bytes.Buffer
	if err := Encode(&output, value, EncodeOptions{Pretty: true, Indent: "  "}); err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"line\": \"a\\n\\tb\",\n  \"values\": [\n    1,\n    true\n  ]\n}"
	if output.String() != want {
		t.Fatalf("pretty:\n%s\nwant:\n%s", output.String(), want)
	}
	if err := Encode(failingWriter{}, value, EncodeOptions{}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("write=%v", err)
	}
	if err := Encode(shortWriter{}, value, EncodeOptions{}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write=%v", err)
	}
	for _, bad := range []Value{{Type: Number, Text: "NaN"}, {Type: String, Text: string([]byte{0xff})}, {Type: Invalid}} {
		if err := Encode(io.Discard, bad, EncodeOptions{}); err == nil {
			t.Fatalf("Encode(%+v) succeeded", bad)
		}
	}
}

func FuzzDecodeDoesNotPanic(f *testing.F) {
	for _, seed := range []string{`null`, `{"a":[1,true,"x"]}`, `"\uD83D\uDE80"`, `[`, string([]byte{0xff})} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = DecodeBytes([]byte(input), Limits{MaxBytes: 4096, MaxDepth: 16, MaxNodes: 256, MaxStringBytes: 1024})
	})
}

func TestEveryScalarAndEscape(t *testing.T) {
	value, err := DecodeBytes([]byte(`[null,true,false,0,-1,1.2,3e4,"\"\\\/\b\f\n\r\t\u0041"]`), Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Elements) != 8 || value.Elements[7].Text != "\"\\/\b\f\n\r\tA" {
		t.Fatalf("value=%+v", value)
	}
}
