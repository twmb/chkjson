package chkjson

import (
	"bytes"
	"encoding/json"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

// This test covers most edge cases of validating and compacting .
func TestMost(t *testing.T) {
	var tests = []string{
		"",
		"   ",
		" z",
		" 1  1",
		" 1  {}",
		" 1  []",
		" 1  true",
		" 1  null",
		" 1  \"n\"",

		// string
		"\"\xe2",       // begin line-sep but invalid finish
		"\"\xe2\x79\"", // begin line-sep but not actually line sep
		`"foo"`,
		"\"\xe2\x80\xa8\xe2\x80\xa9\"", // line-sep and paragraph-sep
		` "\uaaaa" `,
		` "\uaaaa\uaaaa" `,
		` "\`,
		` "\z`,
		" \"f\x00o\"",
		` "foo`,
		` "\uazaa" `,

		// number
		"1",
		"  0 ",
		" 0e1 ",
		"1.",
		"1z",
		" 0e+0 ",
		" -0e+0 ",
		"-0",
		"1e6",
		"1e+6",
		"-1e+6",
		"-0e+6",
		" -103e+1 ",
		"-0.01e+006",
		"-z",
		"-",
		"1e",
		"1e+",
		" 03e+1 ",
		" 1e.1 ",
		" 00 ",
		"1.e3",
		"01e+6",
		"-0.01e+0.6",

		// object
		"{}",
		`{"foo": 3}`,
		` {}    `,
		strings.Repeat(`{"f":`, 1000) + "{}" + strings.Repeat("}", 1000),
		`{"foo": [{"":3, "4": "3"}, 4, {}], "t_wo": 1}`,
		` {"foo": 2,"fudge}`,
		`{{"foo": }}`,
		`{"foo": true, f "a": true}`,
		`{{"foo": [{"":3, 4: "3"}, 4, "5": {4}]}, "t_wo": 1}`,
		`{"\uaaaa\uaaaa" : true}`,
		"{\"\xe2\x80\xa8\xe2\x80\xa9\": true}", // line-sep and paragraph-sep
		"{",
		`{"foo"`,
		`{"foo",f}`,
		`{"foo",`,
		`{"foo"f`,
		"{}}",

		// array
		`[]`,
		`[ ]`,
		`[ 1, {}]`,
		strings.Repeat("[", 1000) + strings.Repeat("]", 1000),
		`[1, 2, 3, 4, {}]`,
		`[`,
		`[1,`,
		`[1a`,
		`[1a]`,
		`[]]`,

		// boolean
		"true",
		"   true ",
		"tru",
		"false",
		"  true f",
		"fals",
		"falsee",

		// null
		"null ",
		" null ",
		"nul",
		" nulll ",
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			// copy the original string to ensure it is not modified in our
			// unsafe code.
			origBytes := []byte(test)
			orig := string(origBytes)

			in := []byte(test)
			val := Valid(in)
			strVal := ValidString(test)
			expVal := json.Valid(in)

			if test != orig {
				t.Errorf("#%d: ValidString modified the original string!", i)
			}
			if val != strVal {
				t.Errorf("«%s»: valid? %v, valid as string? %v", test, val, strVal)
			}
			if val != expVal {
				t.Errorf("«%s»: got valid? %v, truth? %v", test, val, expVal)
			}

			pact, pactVal := AppendCompact(nil, in)
			pactStr, pactValStr := AppendCompactString(nil, test)

			if test != orig {
				t.Errorf("#%d: AppendCompactString modified the original string!", i)
			}
			if pactVal != expVal {
				t.Errorf("«%s»: compact got valid? %v, truth? %v", test, pactVal, expVal)
			}
			if pactVal != pactValStr {
				t.Errorf("«%s»: compact got valid? %v, valid as string? %v", test, pactVal, pactValStr)
			}

			pactInplace, pactInplaceOk := Compact(append([]byte(nil), in...))
			if pactInplaceOk != expVal {
				t.Errorf("«%s»: compact inplace got valid? %v, truth? %v", test, pactInplaceOk, expVal)
			}
			if expVal && pactInplaceOk && !bytes.Equal(pactInplace, pact) {
				t.Errorf("«%s»: compact inplace got «%s» != AppendCompact as inplace «%s»", test, pactInplace, pact)
			}

			if !expVal {
				return
			}

			// If this is valid JSON, check that our compact calls compact
			// appropriately.
			buf := new(bytes.Buffer)
			json.Compact(buf, in)
			expPact := buf.Bytes()

			if !bytes.Equal(pact, expPact) {
				t.Errorf("«%s»: compact got «%s», exp «%s»", test, pact, expPact)
			}
			if !bytes.Equal(pactStr, expPact) {
				t.Errorf("«%s»: compact string got «%s», exp «%s»", test, pactStr, expPact)
			}
		})
	}
}

func BenchmarkValid(b *testing.B) {
	in := []byte(`{"foo": 1, "bar": [{"fi\uabcdrst": 1,  "se\\cond": 2, "last": 9999}, {}]}`)
	if !Valid(in) {
		b.Fatal("benchmark JSON is not valid!")
	}
	for i := 0; i < b.N; i++ {
		Valid(in)
	}
}

func TestValidRand(t *testing.T) {
	for i := 0; i < 10; i++ {
		b := genBig()
		if !Valid(b) {
			t.Error("unexpected invalid")
		}
	}
}

// The gen code below is taken directly from Go:
//
// https://github.com/golang/go/blob/5d11838/src/encoding/json/scanner_test.go
//
// BSD license that governs the file:
//
// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

func genBig() []byte {
	n := 10000
	if testing.Short() {
		n = 100
	}
	b, err := json.Marshal(genValue(n))
	if err != nil {
		panic(err)
	}
	return b
}

func genValue(n int) interface{} {
	if n > 1 {
		switch rand.Intn(2) {
		case 0:
			return genArray(n)
		case 1:
			return genMap(n)
		}
	}
	switch rand.Intn(3) {
	case 0:
		return rand.Intn(2) == 0
	case 1:
		return rand.NormFloat64()
	case 2:
		return genString(30)
	}
	panic("unreachable")
}

func genString(stddev float64) string {
	n := int(math.Abs(rand.NormFloat64()*stddev + stddev/2))
	c := make([]rune, n)
	for i := range c {
		f := math.Abs(rand.NormFloat64()*64 + 32)
		if f > 0x10ffff {
			f = 0x10ffff
		}
		c[i] = rune(f)
	}
	return string(c)
}

func genArray(n int) []interface{} {
	f := int(math.Abs(rand.NormFloat64()) * math.Min(10, float64(n/2)))
	if f > n {
		f = n
	}
	if f < 1 {
		f = 1
	}
	x := make([]interface{}, f)
	for i := range x {
		x[i] = genValue(((i+1)*n)/f - (i*n)/f)
	}
	return x
}

func genMap(n int) map[string]interface{} {
	f := int(math.Abs(rand.NormFloat64()) * math.Min(10, float64(n/2)))
	if f > n {
		f = n
	}
	if n > 0 && f == 0 {
		f = 1
	}
	x := make(map[string]interface{})
	for i := 0; i < f; i++ {
		x[genString(10)] = genValue(((i+1)*n)/f - (i*n)/f)
	}
	return x
}
