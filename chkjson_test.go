package chkjson

import (
	"encoding/json"
	"math"
	"math/rand"
	"strings"
	"testing"
)

// This test covers validating all function calls and most of their edge cases.
func TestMost(t *testing.T) {
	var typTests = []struct {
		in       string
		expValid bool
	}{
		{"", false},
		{"   ", false},
		{" z", false},

		// string
		{`"foo"`, true},
		{"\"\xe2\x80\xa8\xe2\x80\xa9\"", true}, // line-sep and paragraph-sep
		{` "\uaaaa" `, true},
		{` "\`, false},
		{` "\z`, false},
		{" \"f\x00o\"", false},
		{` "foo`, false},
		{` "\uazaa" `, false},

		// number
		{"1", true},
		{"  0 ", true},
		{" 0e1 ", true},
		{" 0e+0 ", true},
		{" -0e+0 ", true},
		{"-0", true},
		{"1e6", true},
		{"1e+6", true},
		{"-1e+6", true},
		{"-0e+6", true},
		{" -103e+1 ", true},
		{"-0.01e+006", true},
		{"-z", false},
		{"-", false},
		{"1e", false},
		{"1e+", false},
		{" 03e+1 ", false},
		{" 1e.1 ", false},
		{" 00 ", false},
		{"1.e3", false},
		{"01e+6", false},
		{"-0.01e+0.6", false},

		// object
		{"{}", true},
		{`{"foo": 3}`, true},
		{` {}    `, true},
		{strings.Repeat(`{"f":`, 1000) + "{}" + strings.Repeat("}", 1000), true},
		{`{"foo": [{"":3, "4": "3"}, 4, {}], "t_wo": 1}`, true},
		{` {"foo": 2,"fudge}`, false},
		{`{{"foo": }}`, false},
		{`{{"foo": [{"":3, 4: "3"}, 4, "5": {4}]}, "t_wo": 1}`, false},
		{"{", false},
		{`{"foo"`, false},
		{`{"foo",f}`, false},
		{`{"foo",`, false},
		{`{"foo"f`, false},
		{"{}}", false},

		// array
		{`[]`, true},
		{strings.Repeat("[", 1000) + strings.Repeat("]", 1000), true},
		{`[1, 2, 3, 4, {}]`, true},
		{`[`, false},
		{`[1,`, false},
		{`[1a`, false},
		{`[]]`, false},

		// boolean
		{"true", true},
		{"   true ", true},
		{"false", true},
		{"  true f", false},
		{"fals", false},
		{"falsee", false},

		// null
		{"null ", true},
		{" null ", true},
		{" nulll ", false},
	}

	for i, test := range typTests {
		in := []byte(test.in)
		isValid := Valid(in)
		truth := json.Valid(in)

		if isValid != truth {
			t.Errorf("#%d «%s»: got valid? %v, truth? %v", i, test.in, isValid, truth)
		}
		if isValid != test.expValid {
			t.Errorf("#%d «%s»: got valid? %v, expected? %v", i, test.in, isValid, test.expValid)
		}
	}
}

func BenchmarkValid(b *testing.B) {
	in := []byte(`{"foo": 1, "bar": [{"first": 1, "second": 2, "last": 9999}, {}]}`)
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
			t.Errorf("«%s» should be valid", b)
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
