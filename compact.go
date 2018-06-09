package chkjson

import (
	"unsafe"
)

// AppendCompact appends the compact form of src to dst if src is entirely
// JSON, returning the updated dst and whether all of src was valid.
//
// This function assumes and returns ownership of dst. If the src contains
// invalid JSON, dst may have an invalid byte.
//
// It is valid to pass (src[:0], src) to this function; src will be overwritten
// with any valid JSON. If src is invalid, the returned slice may have an
// invalid byte. This function is guaranteed not to add bytes to the input
// JSON, meaning this function will not reallocate the input slice if it is
// used as both src and dst.
//
// This function does not escape line-separator or paragraph-separator
// characters, which can be problematic for JSONP. If conversion is necessary,
// use AppendCompactJSONP.
func AppendCompact(dst, src []byte) ([]byte, bool) {
	p := compactor{parser: parser{in: src}, dst: dst}
	return p.dst, p.compact()
}

// AppendCompactJSONP is similar to AppendCompact, but this function escapes
// line-separator and paragraph-separator characters. See the documentation of
// AppendCompact for more details.
//
// It is still valid to pass (src[:0], src) to this function even though the
// six escaped characters are longer than the original unescaped three. If the
// encoding cannot fit three new characters into itself without overwriting its
// src position, it will reallocate a new slice and eventually return that.
//
// This function returns the new or the updated slice and whether the input
// source was valid JSON.
func AppendCompactJSONP(dst, src []byte) ([]byte, bool) {
	p := compactor{parser: parser{in: src}, jsonp: true, dst: dst}
	return p.dst, p.compact()
}

// compactor appends to dst as we parse non-space bytes. Most of the parsing
// code is copied verbatim.
//
// The general pattern here is to keep a byte in dst whenever it is
// non-whitespace while parsing.
//
// While it is tedious to have this code copied, it allows us great perf.
type compactor struct {
	parser

	jsonp bool

	dst []byte
}

func (p *compactor) compact() bool {
	var c byte
	for p.at < len(p.in) {
		c = p.in[p.at]
		if !isSpace(c) {
			goto start
		}
		p.at++
	}
	return false

start:
	for p.at < len(p.in) { // afterSpace
		c = p.in[p.at]
		p.at++
		if !isSpace(c) {
			p.dst = append(p.dst, c)
			goto beginVal
		}
	}
	return p.stack.empty()

beginVal:
	switch c {
	case '{':
		if p.stack.end < 32 { // push
			p.stack.base[p.stack.end] = parseObjKey
			p.stack.end++
		} else {
			p.stack.ext = append(p.stack.ext, parseObjKey)
		}
		goto beginStrOrEmpty

	case '[':
		if p.stack.end < 32 { // push
			p.stack.base[p.stack.end] = parseArrVal
			p.stack.end++
		} else {
			p.stack.ext = append(p.stack.ext, parseArrVal)
		}
		goto beginValOrEmpty

	case '"':
		goto finStr

	case '-':
		goto finNeg

	case '0':
		goto fin0

	case 't':
		goto finTrue

	case 'f':
		goto finFalse

	case 'n':
		goto finNull
	}
	if isNat(c) {
		goto fin1
	}
	return false

beginStrOrEmpty:
	for p.at < len(p.in) { // afterSpace
		c = p.in[p.at]
		p.at++
		if !isSpace(c) {
			break
		}
	}
	p.dst = append(p.dst, c)
	if c == '}' {
		l := len(p.stack.ext)
		if l == 0 {
			p.stack.end--
		} else {
			p.stack.ext = p.stack.ext[:l-1]
		}
		goto endVal
	}
	if c == '"' {
		goto finStr
	}
	return false

beginValOrEmpty:
	for p.at < len(p.in) {
		c = p.in[p.at]
		if !isSpace(c) {
			if c == ']' {
				goto endVal
			}
			goto start
		}
		p.at++
	}
	return false

finStr:
	if p.at >= len(p.in) {
		return false
	}
	c = p.in[p.at]
	p.at++
	if c == '"' {
		p.dst = append(p.dst, c)
		goto endVal
	}
	if c == '\\' {
		p.dst = append(p.dst, c)

		if p.at >= len(p.in) {
			return false
		}
		c = p.in[p.at]
		p.at++
		p.dst = append(p.dst, c) // pre-add the escaped byte

		switch c {
		case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
			goto finStr
		case 'u':
			if p.at+3 < len(p.in) {
				c1 := p.in[p.at]
				c2 := p.in[p.at+1]
				c3 := p.in[p.at+2]
				c4 := p.in[p.at+3]
				if isHex(c1) && isHex(c2) && isHex(c3) && isHex(c4) {
					p.dst = append(p.dst, c1, c2, c3, c4)
					p.at += 4
					goto finStr
				}
			}
		}
		return false

	}
	if c < 0x20 {
		return false
	}
	if !p.jsonp || c != 0xe2 {
		p.dst = append(p.dst, c)
		goto finStr
	}

	// encoding/json's Compact changes U+2028 and U+2029 (0xe280a8 and
	// 0xe280a9) to "\u2028" and "\u2029".
	//
	// If you notice, that turns 3 characters into 6. We provide the option
	// to use src[:0] as dst, but encoding this blindly would clobber what
	// we have yet to read.
	//
	// If this is e280a{8,9}, we set dst to be a new slice unless we have
	// three spare bytes saved from compacting.
	if p.at+1 >= len(p.in) {
		p.dst = append(p.dst, c)
		goto finStr
	}

	if n1, n2 := p.in[p.at], p.in[p.at+1]; n1 != 0x80 || n2&^1 != 0xa8 {
		p.dst = append(p.dst, c)
		goto finStr
	} else {
		// dstPtr, at most, can be just before srcPtr. We have not
		// appended the new character yet.
		//
		// We need to reallocate if our three new characters will put
		// us on or after srcPtr.
		//
		// This conditional is very ugly to avoid declaration problems
		// with goto.
		if srcPtr, dstPtr :=
			uintptr(unsafe.Pointer(&p.in[p.at-1])), // sub one to get where we _were_
			uintptr(unsafe.Pointer(&p.dst[len(p.dst)-1])); dstPtr < srcPtr && dstPtr+3 >= srcPtr {

			new := make([]byte, len(p.dst), cap(p.in))
			copy(new, p.dst)
			p.dst = new
		}

		p.dst = append(p.dst, `\u202`...)
		p.dst = append(p.dst, ('8' | n2&1))
		p.at += 2

		goto finStr
	}

finNeg:
	if p.at >= len(p.in) {
		return false
	}
	c = p.in[p.at]
	p.at++
	p.dst = append(p.dst, c)
	if c == '0' {
		goto fin0
	}
	if isNat(c) {
		goto fin1
	}
	return false

fin1:
	for p.at < len(p.in) {
		c = p.in[p.at]
		if !isNum(c) {
			goto fin0
		}
		p.dst = append(p.dst, c)
		p.at++
	}
	goto endVal

fin0:
	if p.at >= len(p.in) {
		goto endVal
	}
	c = p.in[p.at]
	if c == '.' {
		p.dst = append(p.dst, c)
		p.at++
		goto finDot
	}
	if isE(c) {
		p.dst = append(p.dst, c)
		p.at++
		goto finE
	}
	goto endVal

finDot:
	if p.at >= len(p.in) {
		return false
	}
	c = p.in[p.at]
	p.at++
	if !isNum(c) { // first char after dot must be num
		return false
	}
	p.dst = append(p.dst, c)

	for p.at < len(p.in) { // consume all nums
		c := p.in[p.at]
		if !isNum(c) {
			break
		}
		p.dst = append(p.dst, c)
		p.at++
	}

	if p.at < len(p.in) {
		c := p.in[p.at]
		if isE(p.in[p.at]) {
			p.dst = append(p.dst, c)
			p.at++
			goto finE
		}
	}
	goto endVal

finE:
	if p.at >= len(p.in) {
		return false
	}
	c = p.in[p.at]
	p.at++

	if c == '+' || c == '-' {
		p.dst = append(p.dst, c) // keep before overwriting c
		if p.at >= len(p.in) {
			return false
		}
		c = p.in[p.at]
		p.at++
	}

	if !isNum(c) { // first after e (and +/-) must be num
		return false
	}
	p.dst = append(p.dst, c)

	for p.at < len(p.in) { // consume all nums
		c := p.in[p.at]
		if !isNum(c) {
			break
		}
		p.dst = append(p.dst, c)
		p.at++
	}
	goto endVal

finTrue:
	if p.at+2 < len(p.in) &&
		p.in[p.at] == 'r' &&
		p.in[p.at+1] == 'u' &&
		p.in[p.at+2] == 'e' {

		p.at += 3
		p.dst = append(p.dst, 'r', 'u', 'e')
		goto endVal
	}
	return false

finFalse:
	if p.at+3 < len(p.in) &&
		p.in[p.at] == 'a' &&
		p.in[p.at+1] == 'l' &&
		p.in[p.at+2] == 's' &&
		p.in[p.at+3] == 'e' {

		p.at += 4
		p.dst = append(p.dst, 'a', 'l', 's', 'e')
		goto endVal
	}
	return false

finNull:
	if p.at+2 < len(p.in) &&
		p.in[p.at] == 'u' &&
		p.in[p.at+1] == 'l' &&
		p.in[p.at+2] == 'l' {

		p.at += 3
		p.dst = append(p.dst, 'u', 'l', 'l')
		goto endVal
	}
	return false

endVal:
	if p.stack.empty() {
		for p.at < len(p.in) {
			c = p.in[p.at]
			p.at++
			if !isSpace(c) {
				return false
			}
		}
		goto start
	}

	for p.at < len(p.in) { // if parseState is not empty, we need another character
		c = p.in[p.at]
		p.at++
		if !isSpace(c) {
			p.dst = append(p.dst, c)
			goto finVal
		}
	}
	return false

finVal:
	switch p.stack.pop() {
	case parseObjKey:
		if c == ':' {
			if p.stack.end < 32 { // push
				p.stack.base[p.stack.end] = parseObjVal
				p.stack.end++
			} else {
				p.stack.ext = append(p.stack.ext, parseObjVal)
			}
			goto start
		}

	case parseObjVal:
		switch c {
		case ',':
			if p.stack.end < 32 { // push
				p.stack.base[p.stack.end] = parseObjKey
				p.stack.end++
			} else {
				p.stack.ext = append(p.stack.ext, parseObjKey)
			}
			for p.at < len(p.in) { // afterSpace
				c = p.in[p.at]
				p.at++
				if isSpace(c) {
					continue
				}
				if c == '"' {
					p.dst = append(p.dst, c)
					goto finStr
				}
				return false
			}
		case '}':
			goto endVal
		}

	case parseArrVal:
		switch c {
		case ',':
			if p.stack.end < 32 { // push
				p.stack.base[p.stack.end] = parseArrVal
				p.stack.end++
			} else {
				p.stack.ext = append(p.stack.ext, parseArrVal)
			}
			goto start
		case ']':
			goto endVal
		}
	}

	return false
}
