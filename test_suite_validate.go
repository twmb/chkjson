// +build none

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/twmb/chkjson"
)

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

// This file checks that all y_ and n_ files in
// github.com/nst/JSONTestSuite/test_parsing
// are validated and compacted properly.
//
// The repo is expected to be in the dir this runs in.
func main() {
	const path = "JSONTestSuite/test_parsing/"
	dirents, err := ioutil.ReadDir(path)
	chk(err)
	for _, dirent := range dirents {
		name := dirent.Name()
		raw, err := ioutil.ReadFile(path + name)
		chk(err)

		gotValid := chkjson.Valid(raw)
		should := strings.HasPrefix(name, "y_")
		shouldNot := strings.HasPrefix(name, "n_")

		if should && !gotValid {
			fmt.Printf("FAIL %s\n", name)
		}
		if shouldNot && gotValid {
			fmt.Printf("FAIL %s\n", name)
		}

		officialValid := json.Valid(raw)
		if officialValid != gotValid {
			fmt.Printf("MISMATCH %s (us: %v, stdlib: %v)\n", name, gotValid, officialValid)
		}

		if !should {
			continue
		}

		var exp interface{}
		chk(json.Unmarshal(raw, &exp))

		for _, fn := range []struct {
			name string
			fn   func([]byte, []byte) ([]byte, bool)
		}{
			{"AppendCompact", chkjson.AppendCompact},
			{"AppendCompactJSONP", chkjson.AppendCompactJSONP},
			{"Compact", func(_, in []byte) ([]byte, bool) { return chkjson.Compact(in) }},
		} {
			dup := append([]byte(nil), raw...)
			dup, _ = fn.fn(dup[:0], dup)

			var got interface{}
			chk(json.Unmarshal(dup, &got))

			if !reflect.DeepEqual(exp, got) {
				fmt.Printf("FAIL %s\n", name)
			}
		}
	}
}
