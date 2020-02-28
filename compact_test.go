package chkjson

import (
	"bytes"
	"encoding/json"
	"testing"
	"unsafe"
)

func TestCompact(t *testing.T) {
	for i := 0; i < 5; i++ {
		buf := genBig()
		orig := make([]byte, len(buf))
		copy(orig, buf)

		// Ensure that compacting already compact JSON does not change
		// anything.
		addr := uintptr(unsafe.Pointer(&buf[0]))
		var is bool
		got, is := AppendCompact(buf[:0], buf)
		if !is {
			t.Error("AppendCompact on known good json is invalid")
		}
		if uintptr(unsafe.Pointer(&got[0])) != addr {
			t.Error("AppendCompact on compact buf to itself changed addresses")
		}
		if len(got) != len(orig) {
			t.Errorf("AppendCompact len %v != exp %v", len(got), len(orig))
		}
		if !bytes.Equal(got, orig) {
			t.Error("AppendCompact changed contents of compact json")
		}

		// Ensure that Compact to itself changes nothing.
		cpy := make([]byte, len(got))
		copy(cpy, got)
		cpy, ok := Compact(cpy)
		if !ok {
			t.Error("Compact of known good json is invalid")
		} else if !bytes.Equal(cpy, got) {
			t.Error("Compact of AppendCompact result != AppendCompact result")
		}

		// Indent that and ensure we end up back with our original.
		indentBuf := new(bytes.Buffer)
		json.Indent(indentBuf, buf, "", "  \t ")
		buf = []byte(indentBuf.String())

		// We copy the checks above but compare lengths/contents
		// against the original known-good compact slice.
		addr = uintptr(unsafe.Pointer(&buf[0]))
		got, is = AppendCompact(buf[:0], buf)
		if !is {
			t.Error("AppendCompact on indented good json is invalid")
		}
		if uintptr(unsafe.Pointer(&buf[0])) != addr {
			t.Error("AppendCompact on large buf to itself changed addresses")
		}
		if len(got) != len(orig) {
			t.Errorf("AppendConcat len %v != exp %v", len(got), len(orig))
		}
		if !bytes.Equal(orig, got) {
			t.Error("AppendConcat of indented json changed original compact")
		}
	}
}

func BenchmarkCompact(b *testing.B) {
	orig := []byte(`{"foo": 1, "bar": [{"fi\uabcdrst": 1,  "se\\cond": 2, "last": 9999}, {}]}`)
	a := make([]byte, 0, len(orig))
	for i := 0; i < b.N; i++ {
		AppendCompact(a[:0], orig)
	}
}

func BenchmarkCompactInplace(b *testing.B) {
	orig := []byte(`{"foo":1,"bar":[{"fi\uabcdrst":1,"se\\cond":2,"last":9999},{}]}`)
	b.Run("baseline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AppendCompact(orig[:0], orig)
		}
	})
	b.Run("direct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Compact(orig)
		}
	})
}
