package chkjson

// Compact compacts a slice in place and returns the updated slice and if the
// slice was valid JSON. If it was invliad, this returns nil.
//
// This is similar to AppendCompact(b[:0], b), but is faster and more intuitive.
func Compact(b []byte) ([]byte, bool) {
	w, r, ok := compact(b, 0, 0)
	if !ok {
		return nil, false
	}

	for ; r < len(b); r++ {
		switch b[r] {
		case '\t', '\n', '\r', ' ':
		default:
			return nil, false
		}
	}
	return b[:w], true
}

func compact(b []byte, w, r int) (int, int, bool) {
	rstart := r
	var c byte
	var ok bool

whitespace:
	if r == len(b) {
		return 0, 0, false
	}

	switch c, r = b[r], r+1; c {
	case ' ', '\r', '\t', '\n':
		rstart++
		goto whitespace
	case '{':
		b[w], w = '{', w+1
		rstart++
		goto finObj
	case '[':
		b[w], w = '[', w+1
		rstart++
		goto finArr
	case '"':
		goto finStr
	case 't':
		end := r + len("rue")
		if end <= len(b) &&
			b[r+2] == 'e' &&
			b[r+1] == 'u' &&
			b[r] == 'r' {
			b[w+3], b[w+2], b[w+1], b[w] = 'e', 'u', 'r', 't'
			return w + 4, r + 3, true
		}
		return 0, 0, false
	case 'f':
		end := r + len("alse")
		if end <= len(b) &&
			b[r+3] == 'e' &&
			b[r+2] == 's' &&
			b[r+1] == 'l' &&
			b[r] == 'a' {
			b[w+4], b[w+3], b[w+2], b[w+1], b[w] = 'e', 's', 'l', 'a', 'f'
			return w + 5, r + 4, true
		}
		return 0, 0, false
	case 'n':
		end := r + len("ull")
		if end <= len(b) &&
			b[r+2] == 'l' &&
			b[r+1] == 'l' &&
			b[r] == 'u' {
			b[w+3], b[w+2], b[w+1], b[w] = 'l', 'l', 'u', 'n'
			return w + 4, r + 3, true
		}
		return 0, 0, false
	case '-':
		goto finNeg
	case '0':
		goto fin0
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		goto fin1
	default:
		return 0, 0, false
	}

finStr:
	for ; r < len(b); r++ {
		switch b[r] {
		default:
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
			30, 31:
			return 0, 0, false
		case '"':
			r++
			w += copy(b[w:], b[rstart:r])
			return w, r, true
		case '\\':
			r++
			if r == len(b) {
				return 0, 0, false
			}
			switch b[r] {
			case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
			case 'u':
				if len(b[r:]) > 5 &&
					isHex(b[r+4]) &&
					isHex(b[r+3]) &&
					isHex(b[r+2]) &&
					isHex(b[r+1]) {
					r += 5
					goto finStr
				}
				return 0, 0, false
			default:
				return 0, 0, false
			}
		}
	}
	return 0, 0, false

finObj:
	for r < len(b) { // finish obj immediately or begin a key
		switch c, r = b[r], r+1; c {
		case ' ', '\r', '\t', '\n':
			rstart++
		case '"':
			goto finObjKey
		case '}':
			b[w] = '}'
			return w + 1, r, true
		default:
			return 0, 0, false
		}
	}

finObjKey:
	for ; r < len(b); r++ { // duplicated above for better jumps
		switch b[r] {
		default:
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
			10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
			30, 31:
			return 0, 0, false
		case '"':
			r++
			w += copy(b[w:], b[rstart:r])
			goto finObjSep
		case '\\':
			r++
			if r == len(b) {
				return 0, 0, false
			}
			switch b[r] {
			case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
			case 'u':
				if len(b[r:]) > 5 &&
					isHex(b[r+4]) &&
					isHex(b[r+3]) &&
					isHex(b[r+2]) &&
					isHex(b[r+1]) {
					r += 5
					goto finObjKey
				}
				return 0, 0, false
			default:
				return 0, 0, false
			}
		}
	}
	return 0, 0, false

finObjSep:
	for r < len(b) {
		switch c, r = b[r], r+1; c {
		case ' ', '\r', '\t', '\n':
		case ':':
			b[w], w = ':', w+1
			goto objAny
		default:
			return 0, 0, false
		}
	}

objAny:
	if w, r, ok = compact(b, w, r); !ok {
		return 0, 0, false
	}

	for r < len(b) {
		switch c, r = b[r], r+1; c {
		case ' ', '\r', '\t', '\n':
		case ',':
			b[w], w, rstart = ',', w+1, r-1
			goto beginStr
		case '}':
			b[w] = '}'
			return w + 1, r, true
		default:
			return 0, 0, false
		}
	}

beginStr:
	for r < len(b) {
		switch c, r = b[r], r+1; c {
		case ' ', '\r', '\t', '\n':
		case '"':
			rstart = r - 1
			goto finObjKey
		default:
			return 0, 0, false
		}
	}
	return 0, 0, false

finArr:
	for r < len(b) {
		switch c = b[r]; c {
		case ' ', '\r', '\t', '\n':
			r++
		case ']':
			b[w] = ']'
			return w + 1, r + 1, true
		default:
			goto arrAny
		}
	}

arrAny:
	if w, r, ok = compact(b, w, r); !ok {
		return 0, 0, false
	}

	for r < len(b) {
		switch c, r = b[r], r+1; c {
		case ' ', '\r', '\t', '\n':
		case ',':
			b[w], w = ',', w+1
			goto arrAny
		case ']':
			b[w] = ']'
			return w + 1, r, true
		default:
			return 0, 0, false
		}
	}

	return 0, 0, false

finNeg:
	if r == len(b) {
		return 0, 0, false
	}
	if c, r = b[r], r+1; c == '0' {
		goto fin0
	}
	if !isNat(c) {
		return 0, 0, false
	}

fin1:
	for ; r < len(b) && isNum(b[r]); r++ {
	}

fin0:
	if r == len(b) {
		w += copy(b[w:], b[rstart:r])
		return w, r, true
	}
	c = b[r]
	if isE(c) {
		r++
		goto finE
	}
	if c != '.' {
		w += copy(b[w:], b[rstart:r])
		return w, r, true
	}
	r++

	// finDot
	if r == len(b) {
		return 0, 0, false
	}
	if c, r = b[r], r+1; !isNum(c) { // first char after dot must be num
		return 0, 0, false
	}

	for ; r < len(b) && isNum(b[r]); r++ {
	}

	if r == len(b) || !isE(b[r]) {
		w += copy(b[w:], b[rstart:r])
		return w, r, true
	}
	r++

finE:
	if r == len(b) {
		return 0, 0, false
	}
	if c, r = b[r], r+1; c == '+' || c == '-' {
		if r == len(b) {
			return 0, 0, false
		}
		c, r = b[r], r+1
	}
	if !isNum(c) { // first after e (and +/-) must be num
		return 0, 0, false
	}
	for ; r < len(b) && isNum(b[r]); r++ {
	}
	w += copy(b[w:], b[rstart:r])
	return w, r, true
}
