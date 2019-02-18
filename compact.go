package chkjson

import (
	"unsafe"
)

// AppendCompact appends the compact form of src to dst if src is entirely
// JSON, returning the updated dst and whether all of src was valid.
//
// This function assumes and returns ownership of dst. If src is invalid, this
// will return nil.
//
// It is valid to pass (src[:0], src) to this function; src will be overwritten
// with any valid JSON. This function is guaranteed not to add bytes to the
// input JSON, meaning this function will not reallocate the input slice if it
// is used as both src and dst.
//
// This function does not escape line-separator or paragraph-separator
// characters, which can be problematic for JSONP. If conversion is necessary,
// use AppendCompactJSONP.
func AppendCompact(dst, src []byte) ([]byte, bool) {
	return AppendCompactString(dst, *(*string)(unsafe.Pointer(&src)))
}

// AppendCompactString is exactly like AppendCompact but for compacting
// strings.
func AppendCompactString(dst []byte, src string) ([]byte, bool) {
	dst, at, ok := packAny(dst, src, 0, false)
	if !ok {
		return nil, false
	}

	for ; at < len(src); at++ {
		switch src[at] {
		case '\t', '\n', '\r', ' ':
		default:
			return nil, false
		}
	}
	return dst, true
}

// AppendCompactJSONP is similar to AppendCompact, but this function escapes
// line-separator and paragraph-separator characters. See the documentation of
// AppendCompact for more details.
//
// It is still valid to pass (src[:0], src) to this function even though the
// six escaped characters replacing line-separator and paragraph-separator
// (each three bytes) are longer than the original unescaped three. If the
// encoding cannot fit three new characters into itself without overwriting its
// src position, it will reallocate a new slice and eventually return that.
//
// This function returns the new or the updated slice and whether the input
// source was valid JSON. If validation fails, this returns nil.
func AppendCompactJSONP(dst, src []byte) ([]byte, bool) {
	return AppendCompactJSONPString(dst, *(*string)(unsafe.Pointer(&src)))
}

// AppendCompactJSONPString is exactly like AppendCompactJSONP but for
// compacting strings.
func AppendCompactJSONPString(dst []byte, src string) ([]byte, bool) {
	dst, at, ok := packAny(dst, src, 0, true)
	if !ok {
		return nil, false
	}

	for ; at < len(src); at++ {
		switch src[at] {
		case '\t', '\n', '\r', ' ':
		default:
			return nil, false
		}
	}
	return dst, true
}

// As in chkjson.go, this function is super ugly. It is mostly a copy of
// chkjson's any but we keep bytes as appropriate.
//
// Where possible, we append spans of memory at a time (via start and at).
// jsonp makes string parsing quite ugly.
func packAny(dst []byte, src string, at int, jsonp bool) ([]byte, int, bool) {
	start := at
	var c byte
	var ok bool

whitespace:
	if at == len(src) {
		return nil, 0, false
	}

	switch c, at = src[at], at+1; c {
	case ' ', '\r', '\t', '\n':
		start++
		goto whitespace
	case '{':
		start++
		dst = append(dst, '{')
		goto finObj
	case '[':
		start++
		dst = append(dst, '[')
		goto finArr
	case '"':
		goto finStr
	case 't':
		end := at + len("rue")
		if end <= len(src) &&
			src[at] == 'r' &&
			src[at+1] == 'u' &&
			src[at+2] == 'e' {
			dst = append(dst, 't', 'r', 'u', 'e')
			return dst, end, true
		}
		return nil, 0, false
	case 'f':
		end := at + len("alse")
		if end <= len(src) && src[at:end] == "alse" {
			dst = append(dst, 'f', 'a', 'l', 's', 'e')
			return dst, end, true
		}
		return nil, 0, false
	case 'n':
		end := at + len("ull")
		if end <= len(src) &&
			src[at] == 'u' &&
			src[at+1] == 'l' &&
			src[at+2] == 'l' {
			dst = append(dst, 'n', 'u', 'l', 'l')
			return dst, end, true
		}
		return nil, 0, false
	case '-':
		goto finNeg
	case '0':
		goto fin0
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		goto fin1
	default:
		return nil, 0, false
	}

finStr:
	for ; at < len(src); at++ {
		switch src[at] {
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
			30, 31:
			return nil, 0, false
		case '"':
			at++
			dst = append(dst, src[start:at]...)
			return dst, at, true
		case '\\':
			at++
			if at == len(src) {
				return nil, 0, false
			}
			switch src[at] {
			case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
			case 'u':
				if len(src[at:]) > 5 &&
					isHex(src[at+1]) &&
					isHex(src[at+2]) &&
					isHex(src[at+3]) &&
					isHex(src[at+4]) {
					at += 5
					goto finStr
				}
				return nil, 0, false
			default:
				return nil, 0, false
			}
		default:
			if !jsonp || src[at] != 0xe2 || at+2 >= len(src) {
				continue
			}

			// encoding/json's Compact changes 0xe280a{8,9} to
			// `\u2028` or `\u2029`, turning 3 bytes into 6. We
			// allow using src[:0] as dst, so we must not clobber
			// what we have yet to read.
			if n1, n2 := src[at+1], src[at+2]; n1 == 0x80 && n2&^1 == 0xa8 {
				dst, start = append(dst, src[start:at]...), at

				overlapMin := len(dst) + 3
				if overlapMin <= cap(dst) {
					overlapTest := dst[len(dst):overlapMin]
					for i := 0; i < len(overlapTest); i++ {
						overlapTest[i] = 0
						if src[at] == 0 { // we just modified src by changing dst; these regions overlap
							new := make([]byte, len(dst), cap(dst))
							copy(new, dst)
							dst = new
							break
						}
					}
				}

				dst = append(dst, '\\', 'u', '2', '0', '2', '8'|n2&1)
				start += 3
				at += 3 // 0xe2, 0x80, 0xa{8,9}

				goto finStr
			}
		}
	}
	return nil, 0, false

finObj:
	for at < len(src) { // finish obj immediately or begin a key
		switch c, at = src[at], at+1; c {
		case ' ', '\r', '\t', '\n':
			start++
		case '"':
			goto finObjKey
		case '}':
			dst = append(dst, '}')
			return dst, at, true
		default:
			return dst, at, false
		}
	}

finObjKey:
	for ; at < len(src); at++ { // duplicated above for better jumps
		switch src[at] {
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
			30, 31:
			return nil, 0, false
		case '"':
			at++
			dst, start = append(dst, src[start:at]...), at
			goto finObjSep
		case '\\':
			at++
			if at == len(src) {
				return nil, 0, false
			}
			switch src[at] {
			case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
			case 'u':
				if len(src[at:]) > 5 &&
					isHex(src[at+1]) &&
					isHex(src[at+2]) &&
					isHex(src[at+3]) &&
					isHex(src[at+4]) {
					at += 5
					goto finObjKey
				}
				return nil, 0, false
			default:
				return nil, 0, false
			}
		default:
			if !jsonp || src[at] != 0xe2 || at+2 >= len(src) {
				continue
			}
			if n1, n2 := src[at+1], src[at+2]; n1 == 0x80 && n2&^1 == 0xa8 {
				dst, start = append(dst, src[start:at]...), at
				overlapMin := len(dst) + 3
				if overlapMin <= cap(dst) {
					overlapTest := dst[len(dst):overlapMin]
					for i := 0; i < len(overlapTest); i++ {
						overlapTest[i] = 0
						if src[at] == 0 {
							new := make([]byte, len(dst), cap(dst))
							copy(new, dst)
							dst = new
							break
						}
					}
				}
				dst = append(dst, '\\', 'u', '2', '0', '2', '8'|n2&1)
				start += 3
				at += 3

				goto finObjKey
			}
		}
	}
	return nil, 0, false

finObjSep:
	for at < len(src) {
		switch c, at = src[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case ':':
			dst = append(dst, ':')
			goto objAny
		default:
			return nil, 0, false
		}
	}

objAny:
	if dst, at, ok = packAny(dst, src, at, jsonp); !ok {
		return nil, 0, false
	}

	for at < len(src) {
		switch c, at = src[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case ',':
			dst, start = append(dst, ','), at-1
			goto beginStr
		case '}':
			dst = append(dst, '}')
			return dst, at, true
		default:
			return nil, 0, false
		}
	}

beginStr:
	for at < len(src) {
		switch c, at = src[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case '"':
			start = at - 1
			goto finObjKey
		default:
			return nil, 0, false
		}
	}
	return nil, 0, false

finArr:
	for at < len(src) {
		switch c = src[at]; c {
		case ' ', '\r', '\t', '\n':
			at++
		case ']':
			dst = append(dst, ']')
			return dst, at + 1, true
		default:
			goto arrAny
		}
	}

arrAny:
	if dst, at, ok = packAny(dst, src, at, jsonp); !ok {
		return nil, 0, false
	}

	for at < len(src) {
		switch c, at = src[at], at+1; c {
		case ' ', '\r', '\t', '\n':
		case ',':
			dst = append(dst, ',')
			goto arrAny
		case ']':
			dst = append(dst, ']')
			return dst, at, true
		default:
			return nil, 0, false
		}
	}

	return nil, 0, false

finNeg:
	if at == len(src) {
		return nil, 0, false
	}
	if c, at = src[at], at+1; c == '0' {
		goto fin0
	}
	if !isNat(c) {
		return nil, 0, false
	}

fin1:
	for ; at < len(src) && isNum(src[at]); at++ {
	}

fin0:
	if at == len(src) {
		dst = append(dst, src[start:at]...)
		return dst, at, true
	}
	c = src[at]
	if isE(c) {
		at++
		goto finE
	}
	if c != '.' {
		dst = append(dst, src[start:at]...)
		return dst, at, true
	}
	at++

	// finDot
	if at == len(src) {
		return nil, 0, false
	}
	if c, at = src[at], at+1; !isNum(c) { // first char after dot must be num
		return nil, 0, false
	}

	for ; at < len(src) && isNum(src[at]); at++ {
	}

	if at == len(src) || !isE(src[at]) {
		dst = append(dst, src[start:at]...)
		return dst, at, true
	}
	at++

finE:
	if at == len(src) {
		return nil, 0, false
	}
	if c, at = src[at], at+1; c == '+' || c == '-' {
		if at == len(src) {
			return nil, 0, false
		}
		c, at = src[at], at+1
	}
	if !isNum(c) { // first after e (and +/-) must be num
		return nil, 0, false
	}
	for ; at < len(src) && isNum(src[at]); at++ {
	}
	dst = append(dst, src[start:at]...)
	return dst, at, true
}
