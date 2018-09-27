package chkjson

import (
	"bytes"
	"testing"
)

func TestEscape(t *testing.T) {
	tests := []struct {
		in   string
		opts []EscapeOpt
		exp  string
	}{
		{"foo", []EscapeOpt{EscapeHTML, EscapeJSONP}, "foo"}, // basic test

		// more advanced tests on the same input
		{
			"\xffz\n\r\t\"\x00\x1f<>\u2028\u2029\u2030&",
			nil,
			"\\ufffdz\\n\\r\\t\\\"\\u0000\\u001f<>\u2028\u2029\u2030&",
		},
		{
			"\xffz\n\r\t\"\x00\x1f<>\u2028\u2029\u2030&",
			[]EscapeOpt{EscapeHTML},
			"\\ufffdz\\n\\r\\t\\\"\\u0000\\u001f\\u003c\\u003e\u2028\u2029\u2030\\u0026",
		},
		{
			"\xffz\n\r\t\"\x00\x1f<>\u2028\u2029\u2030&",
			[]EscapeOpt{EscapeJSONP},
			"\\ufffdz\\n\\r\\t\\\"\\u0000\\u001f<>\\u2028\\u2029\u2030&",
		},
		{
			"\xffz\n\r\t\"\x00\x1f<>\u2028\u2029\u2030&",
			[]EscapeOpt{EscapeJSONP, EscapeHTML},
			"\\ufffdz\\n\\r\\t\\\"\\u0000\\u001f\\u003c\\u003e\\u2028\\u2029\u2030\\u0026",
		},
	}

	for i, test := range tests {
		gotB := Escape(nil, []byte(test.in), test.opts...)
		got := EscapeString(nil, test.in, test.opts...)
		if !bytes.Equal(gotB, got) {
			t.Errorf("#%d: got bytes != got string", i)
		}
		if string(got) != test.exp {
			t.Errorf("#%d: got %+q != exp %+q", i, got, test.exp)
		}
	}
}

func BenchmarkEscapeEasy(b *testing.B) {
	benchmarkEscape(b, "aaaaaaaaaaaaaaa")
}

func BenchmarkEscapeHard(b *testing.B) {
	benchmarkEscape(b, "\xffz\n\r\t\"\x00\x1f<>\u2028\u2029\u2030&")
}

func benchmarkEscape(b *testing.B, ins string) {
	in := []byte(ins)
	out := make([]byte, 0, 100)

	for _, combo := range []struct {
		name string
		opts []EscapeOpt
	}{
		{"no_opts", nil},
		{"html", []EscapeOpt{EscapeHTML}},
		{"jsonp", []EscapeOpt{EscapeJSONP}},
		{"both", []EscapeOpt{EscapeHTML, EscapeJSONP}},
	} {
		b.Run(combo.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Escape(out, in, combo.opts...)
			}
		})
	}
}
