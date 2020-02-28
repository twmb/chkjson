package chkjson

import (
	"fmt"
	"io/ioutil"
	"testing"
)

var extFiles = make(map[string][]byte)

func init() {
	for _, fname := range []string{
		"small", "medium", "large", "canada", "citm", "twitter",
	} {
		path := "testdata/" + fname + ".json"
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("unable to read %s: %v", path, err))
		}
		extFiles[fname] = bs
	}
}

func TestExtValid(t *testing.T) {
	for fname, bs := range extFiles {
		t.Run(fname, func(t *testing.T) {
			if !Valid(bs) {
				t.Errorf("%s unexpectedly invalid!", fname)
			}
		})
	}
}

func TestExtAppendCompact(t *testing.T) {
	for fname, bs := range extFiles {
		t.Run(fname, func(t *testing.T) {
			got, valid := AppendCompact(nil, bs)
			if !valid {
				t.Errorf("%s unexpectedly invalid!", fname)
			}
			if !Valid(got) {
				t.Errorf("%s compacted invalid!", fname)
			}
		})
	}
}

func TestExtAppendCompactJSON(t *testing.T) {
	for fname, bs := range extFiles {
		t.Run(fname, func(t *testing.T) {
			got, valid := AppendCompact(nil, bs)
			if !valid {
				t.Errorf("%s unexpectedly invalid!", fname)
			}
			if !Valid(got) {
				t.Errorf("%s compacted invalid!", fname)
			}
		})
	}
}

func BenchmarkExtValid(b *testing.B) {
	for fname, bs := range extFiles {
		b.Run(fname, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(bs)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Valid(bs)
			}
		})
	}
}

func BenchmarkExtCompact(b *testing.B) {
	for fname, bs := range extFiles {
		b.Run(fname, func(b *testing.B) {
			buf, _ := AppendCompact(nil, bs)
			b.ReportAllocs()
			b.SetBytes(int64(len(bs)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				AppendCompact(buf[:0], bs)
			}
		})
	}
}

func BenchmarkExtCompactInplace(b *testing.B) {
	for fname, bs := range extFiles {
		b.Run(fname, func(b *testing.B) {
			buf, _ := AppendCompact(nil, bs)
			b.ReportAllocs()
			b.SetBytes(int64(len(bs)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Compact(buf)
			}
		})
	}
}

/* if you are curious for comparison, uncomment below.
func BenchmarkExtValidStdlib(b *testing.B) {
	for fname, bs := range extFiles {
		b.Run(fname, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(bs)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				json.Valid(bs)
			}
		})
	}
}

func BenchmarkExtCompactStdlib(b *testing.B) {
	for fname, bs := range extFiles {
		b.Run(fname, func(b *testing.B) {
			buf := new(bytes.Buffer)
			buf.Write(bs)
			b.ReportAllocs()
			b.SetBytes(int64(len(bs)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				json.Compact(buf, bs)
			}
		})
	}
}
*/
