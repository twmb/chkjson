// +build gofuzz

package chkjson

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func Fuzz(data []byte) int {
	got := Valid(data)
	exp := json.Valid(data)
	if got != exp {
		panic(fmt.Sprintf("got? %v, exp? %v", got, exp))
	}

	if got != true { // not valid
		if _, compacted := AppendCompact(data[:0], data); compacted {
			panic("compacted invalid json!")
		}
		return 0
	}

	jb := make([]byte, 0, len(data))
	b := bytes.NewBuffer(jb)
	if err := json.Compact(b, data); err != nil {
		panic(fmt.Sprintf("invalid stdlib: %v!", err))
	}
	compact1, ok1 := AppendCompact(nil, data)
	compact2, ok2 := AppendCompactJSONP(data[:0], data)

	if !ok1 {
		panic("compact valid to nil, not ok!")
	}
	if !ok2 {
		panic("compact valid to self jsonp, not ok!")
	}

	if !bytes.Equal(b.Bytes(), compact2) {
		panic("not equal stdlib and compact jsonp!")
	}

	if len(compact1) > len(compact2) {
		panic("compact without jsonp larger than with jsonp!")
	}

	if !Valid(compact1) {
		panic("compact without jsonp not valid!")
	}

	esc := Escape(jb[:0], data, EscapeHTML, EscapeJSONP)
	canonEsc, _ := json.Marshal(string(data))
	canonEsc = canonEsc[1 : len(canonEsc)-1]
	if !bytes.Equal(esc, canonEsc) {
		panic("escape html+jsonp != canonical marshal!")
	}

	return 1
}
