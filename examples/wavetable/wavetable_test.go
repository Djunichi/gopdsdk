package wavetable

import (
	"encoding/binary"
	"testing"
)

func TestSawTable(t *testing.T) {
	v := sawTable(256)
	if len(v) != 512 {
		t.Fatal(len(v))
	}
	if int16(binary.LittleEndian.Uint16(v[:2])) > -32000 {
		t.Fatal("missing negative edge")
	}
	if int16(binary.LittleEndian.Uint16(v[510:])) < 32000 {
		t.Fatal("missing positive edge")
	}
}
