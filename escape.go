package chkjson

import (
	"unicode/utf8"
	"unsafe"
)

// Most of this file is taken basically from Go's encoding/json.encodeState's
// string function. Given that, this file is governed by Go's BSD license.
// See the license in the middle of chkjson_test.

// EscapeOpt is a type for escaping.
type EscapeOpt int

const (
	// EscapeHTML causes Escape to ensure that all <, >, and & are safely
	// HTML escaped.
	EscapeHTML EscapeOpt = iota
	// EscapeJSONP causes Escape to ensure that the line separator and
	// paragraph separator unicode characters are safely escaped for JSONP.
	EscapeJSONP
)

// Escape appends a JSON escaped src to dst.
//
// This function takes options to configure additional escaping.
//
// It is not to use src as dst; there is no check to see if they overlap.
func Escape(dst, src []byte, opts ...EscapeOpt) []byte {
	return EscapeString(dst, *(*string)(unsafe.Pointer(&src)), opts...)
}

// EscapeString appends a JSON escaped src to dst.
//
// This is the same as Escape, but for strings.
func EscapeString(dst []byte, src string, opts ...EscapeOpt) []byte {
	const hex = "0123456789abcdef"
	var html, jsonp bool
	for _, opt := range opts {
		switch opt {
		case EscapeHTML:
			html = true
		case EscapeJSONP:
			jsonp = true
		}
	}

	st := 0
	for i := 0; i < len(src); { // i incremented manually
		if c := src[i]; c < utf8.RuneSelf {
			if htmlSafeSet[c] || (!html && safeSet[c]) {
				i++
				continue
			}
			dst = append(dst, src[st:i]...)
			switch c {
			case '"', '\\':
				dst = append(dst, '\\', c)
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			default:
				dst = append(dst, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xf])
			}
			i++
			st = i
			continue
		}

		c, sz := utf8.DecodeRuneInString(src[i:])
		if c == utf8.RuneError && sz == 1 {
			dst = append(dst, src[st:i]...)
			dst = append(dst, `\ufffd`...)
			i++
			st = i
			continue
		}

		if (c == '\u2028' || c == '\u2029') && jsonp {
			dst = append(dst, src[st:i]...)
			dst = append(dst, '\\', 'u', '2', '0', '2', hex[c&0xf])
			i += sz
			st = i
			continue
		}
		i += sz
	}
	return append(dst, src[st:]...)
}

var safeSet = [utf8.RuneSelf]bool{
	' ': true, '!': true, '#': true, '$': true, '%': true, '&': true,
	'\'': true, '(': true, ')': true, '*': true, '+': true, ',': true,
	'-': true, '.': true, '/': true, '0': true, '1': true, '2': true,
	'3': true, '4': true, '5': true, '6': true, '7': true, '8': true,
	'9': true, ':': true, ';': true, '<': true, '=': true, '>': true,
	'?': true, '@': true, 'A': true, 'B': true, 'C': true, 'D': true,
	'E': true, 'F': true, 'G': true, 'H': true, 'I': true, 'J': true,
	'K': true, 'L': true, 'M': true, 'N': true, 'O': true, 'P': true,
	'Q': true, 'R': true, 'S': true, 'T': true, 'U': true, 'V': true,
	'W': true, 'X': true, 'Y': true, 'Z': true, '[': true, ']': true,
	'^': true, '_': true, '`': true, 'a': true, 'b': true, 'c': true,
	'd': true, 'e': true, 'f': true, 'g': true, 'h': true, 'i': true,
	'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true,
	'p': true, 'q': true, 'r': true, 's': true, 't': true, 'u': true,
	'v': true, 'w': true, 'x': true, 'y': true, 'z': true, '{': true,
	'|': true, '}': true, '~': true, '\u007f': true,
}

var htmlSafeSet = [utf8.RuneSelf]bool{
	' ': true, '!': true, '#': true, '$': true, '%': true, '\'': true,
	'(': true, ')': true, '*': true, '+': true, ',': true, '-': true,
	'.': true, '/': true, '0': true, '1': true, '2': true, '3': true,
	'4': true, '5': true, '6': true, '7': true, '8': true, '9': true,
	':': true, ';': true, '=': true, '?': true, '@': true, 'A': true,
	'B': true, 'C': true, 'D': true, 'E': true, 'F': true, 'G': true,
	'H': true, 'I': true, 'J': true, 'K': true, 'L': true, 'M': true,
	'N': true, 'O': true, 'P': true, 'Q': true, 'R': true, 'S': true,
	'T': true, 'U': true, 'V': true, 'W': true, 'X': true, 'Y': true,
	'Z': true, '[': true, ']': true, '^': true, '_': true, '`': true,
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true,
	'g': true, 'h': true, 'i': true, 'j': true, 'k': true, 'l': true,
	'm': true, 'n': true, 'o': true, 'p': true, 'q': true, 'r': true,
	's': true, 't': true, 'u': true, 'v': true, 'w': true, 'x': true,
	'y': true, 'z': true, '{': true, '|': true, '}': true, '~': true,
	'\u007f': true,
}
