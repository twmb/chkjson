package chkjson

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"unsafe"
)

func TestCompactSeparators(t *testing.T) {
	tests := []struct {
		in, exp string
	}{
		{"{\"\u2028\":1}", `{"\u2028":1}`},
		{"{\"\u2029\":2}", `{"\u2029":2}`},
	}
	for _, test := range tests {
		for i := 0; i < 10; i++ {
			in := []byte(strings.Repeat(" ", i) + test.in)
			realloc := i <= 2

			inaddr := uintptr(unsafe.Pointer(&in[0]))

			gotOrig, ok := AppendCompact(nil, in)
			if !ok {
				t.Error("AppendCompact: unexpected not ok")
			}

			if !bytes.Equal(gotOrig, []byte(test.in)) {
				t.Errorf("AppendCompact: got %q != exp %q", gotOrig, test.in)
			}

			got, ok := AppendCompactJSONP(in[:0], in)

			if !ok {
				t.Fatal("AppendCompactJSONP: unexpected not ok")
			}

			gotaddr := uintptr(unsafe.Pointer(&got[0]))
			if realloc {
				if inaddr == gotaddr {
					t.Errorf("expected realloc with %d leading spaces", i)
				}
			} else {
				if inaddr != gotaddr {
					t.Errorf("unexpected realloc with %d leading spaces", i)
				}
			}

			if !bytes.Equal(got, []byte(test.exp)) {
				t.Errorf("got %q != exp %q", got, test.exp)
			}
		}
	}
}

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
