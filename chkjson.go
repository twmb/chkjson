// Package chkjson provides the fastest JSON validating, compacting, and
// escaping functions that exist for Go.
//
// The standard library allocates when validating JSON and allocates more if
// the JSON is very nested. This library avoids allocating by using recursive
// parsing, which only adds just over 2x memory overhead. Other minor to major
// differences allow this library's Valid to over 2x faster as encoding/json's.
//
// The standard library ensures that compact JSON is JavaScript safe. This is
// necessary if the JSON will ever end up in JSONP, but is not always
// necessary. This library provides a faster AppendCompactJSONP function to
// imitate the stdlib's Compact, and further provides AppendCompact for the
// same speed with a no-allocation (no JSONP escaping) guarantee. This
// library's AppendCompact is usually around 3x-4x faster than encoding/json's
// Compact.
//
// In essence, this library aims to provide faster and allocation free
// alternatives to encoding/json for a few specific use cases. For use cases
// and design considerations, visit this project's repo's README.
package chkjson

import (
	"unsafe"
)

// Valid validates JSON starting from the knowledge that some non-space
// character exists in the parser.
//
// This validation is modelled after the stdlib's validation but operates
// slightly differently. The stdlib validates loops over every character and
// runs the step function for that character. If successful, the step function
// sets the next step function; if not, it sets an error.
//
// Using this step function pattern forces the scanner that holds the step
// function to be allocated: the scanner holds a func(*scanner) field, and when
// passing itself to the func, escape analysis cannot assume that no func will
// escape the scanner. Thus, it has to allocate.
//
// Prior iterations of this code used a stack based scanner similar to
// encoding/json. Recursion is faster and only ~2x more memory expensive.
//
// We use a giant state-machine function and inline many functions for
// significant performance gains.

// Valid returns whether b is valid JSON.
func Valid(b []byte) bool {
	return ValidString(*(*string)(unsafe.Pointer(&b)))
}

// ValidString returns whether s is valid JSON.
func ValidString(s string) bool {
	at, ok := any(s, 0)
	if !ok {
		return false
	}

	for ; at < len(s); at++ {
		switch s[at] {
		case '\t', '\n', '\r', ' ':
		default:
			return false
		}
	}
	return true
}

func any(in string, at int) (int, bool) {
	var c byte
	var ok bool
start:
	if at == len(in) {
		return at, false
	}

	switch c, at = in[at], at+1; c {
	case ' ', '\r', '\t', '\n':
		goto start
	case '{':
		goto finObj
	case '[':
		goto finArr
	case '"':
		goto finStr
	case 't':
		end := at + len("rue")
		return end, end <= len(in) && in[at:end] == "rue"
	case 'f':
		end := at + len("alse")
		return end, end <= len(in) && in[at:end] == "alse"
	case 'n':
		end := at + len("ull")
		return end, end <= len(in) && in[at:end] == "ull"
	case '-':
		goto finNeg
	case '0':
		goto fin0
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		goto fin1
	default:
		return at, false
	}

finStr:
	for ; at < len(in); at++ {
		switch in[at] {
		default:
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
			30, 31:
			return at, false
		case '"':
			return at + 1, true
		case '\\':
			at++
			if at == len(in) {
				return at, false
			}
			switch in[at] {
			case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
			case 'u':
				if len(in[at:]) > 5 &&
					isHex(in[at+1]) &&
					isHex(in[at+2]) &&
					isHex(in[at+3]) &&
					isHex(in[at+4]) {
					at += 5
					goto finStr
				}
				return at, false
			default:
				return at, false
			}
		}
	}
	return at, false

finObj:
	for at < len(in) { // finish obj immediately or begin a key
		switch c, at = in[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case '"':
			goto finObjKey
		case '}':
			return at, true
		default:
			return at, false
		}
	}

finObjKey: // we duplicate the above finStr for better jumps
	for ; at < len(in); at++ {
		switch in[at] {
		default:
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
			30, 31:
			return at, false
		case '"':
			at++
			goto finObjSep
		case '\\':
			at++
			if at == len(in) {
				return at, false
			}
			switch in[at] {
			case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
			case 'u':
				if len(in[at:]) > 5 &&
					isHex(in[at+1]) &&
					isHex(in[at+2]) &&
					isHex(in[at+3]) &&
					isHex(in[at+4]) {
					at += 5
					goto finObjKey
				}
				return at, false
			default:
				return at, false
			}
		}
	}
	return at, false

finObjSep:
	for at < len(in) { // found key, look for colon
		switch c, at = in[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case ':':
			goto objAny
		default:
			return at, false
		}
	}

objAny:
	if at, ok = any(in, at); !ok { // found colon, finish anything
		return at, false
	}

	for at < len(in) { // either end obj or require another key with comma
		switch c, at = in[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case ',':
			goto beginStr
		case '}': // ended obj
			return at, true
		default:
			return at, false
		}
	}

beginStr:
	for at < len(in) { // found comma, look for key beginning
		switch c, at = in[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case '"': // began str
			goto finObjKey
		default:
			return at, false
		}
	}
	return at, false

finArr:
	for at < len(in) { // finish arr immediately or begin anything
		switch c = in[at]; c {
		case ' ', '\r', '\t', '\n':
			at++
		case ']':
			return at + 1, true
		default:
			goto arrAny
		}
	}

arrAny:
	if at, ok = any(in, at); !ok {
		return at, false
	}

	for at < len(in) { // either see a comma and require another anything, or finish
		switch c, at = in[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case ',':
			goto arrAny
		case ']':
			return at, true
		default:
			return at, false
		}
	}

	return at, false

finNeg:
	if at == len(in) {
		return at, false
	}
	if c, at = in[at], at+1; c == '0' {
		goto fin0
	}
	if !isNat(c) {
		return at, false
	}

fin1:
	for ; at < len(in) && isNum(in[at]); at++ {
	}

fin0:
	if at == len(in) {
		return at, true
	}
	c = in[at]
	if isE(c) {
		at++
		goto finE
	}
	if c != '.' {
		return at, true
	}
	at++

	// finDot
	if at == len(in) {
		return at, false
	}
	if c, at = in[at], at+1; !isNum(c) { // first char after dot must be num
		return at, false
	}

	for ; at < len(in) && isNum(in[at]); at++ {
	}

	if at == len(in) || !isE(in[at]) {
		return at, true
	}
	at++

finE:
	if at == len(in) {
		return at, false
	}
	if c, at = in[at], at+1; c == '+' || c == '-' {
		if at == len(in) {
			return at, false
		}
		c, at = in[at], at+1
	}
	if !isNum(c) { // first after e (and +/-) must be num
		return at, false
	}
	for ; at < len(in) && isNum(in[at]); at++ {
	}
	return at, true
}

func isHex(c byte) bool {
	// This switch compiles best. We can recreate it with comparisons directly
	// but we would have to know the proper order.
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'a', 'b', 'c', 'd', 'e', 'f',
		'A', 'B', 'C', 'D', 'E', 'F':
		return true
	default:
		return false
	}
}

func isNum(c byte) bool {
	// With good branch prediction, this in switch form is one cycle
	// faster. In the normal case, we'll have a run of numbers until we
	// don't.
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

func isNat(c byte) bool {
	switch c {
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

func isE(c byte) bool {
	return c == 'e' || c == 'E'
}
