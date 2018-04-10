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
	_, ok := p.peekAfterSpace()
	if !ok {
		return dst, false
	}
	return p.dst, p.run()
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
	_, ok := p.peekAfterSpace()
	if !ok {
		return dst, false
	}
	return p.dst, p.run()
}

// compactor appends to dst as we parse non-space bytes. Most of the parsing
// code is copied verbatim.
//
// The general pattern here is to call p.keep whenever we take a byte from the
// internal parser with any of p.next, p.afterSpace, p.peek+p.skip, or p.tryN.
//
// If this package's code were more general (i.e. like encoding/json), we could
// avoid the duplication, but that would either force us to allocate or force a
// huge slowdown from an unnecessary switch statement.
type compactor struct {
	parser

	jsonp bool

	dst []byte
}

func (p *compactor) keep(c byte) { // called after both next and afterSpace
	p.dst = append(p.dst, c)
}

func (p *compactor) trykeep3(c1, c2, c3 byte) bool {
	if p.try3(c1, c2, c3) {
		p.dst = append(p.dst, c1, c2, c3)
		return true
	}
	return false
}

func (p *compactor) trykeep4(c1, c2, c3, c4 byte) bool {
	if p.try4(c1, c2, c3, c4) {
		p.dst = append(p.dst, c1, c2, c3, c4)
		return true
	}
	return false
}

func (p *compactor) keepskip() { // for keep+skip
	p.dst = append(p.dst, p.on())
	p.skip()
}

func (p *compactor) run() bool {
	var c byte
	ok := true
	for ok {
		if c, ok = p.afterSpace(); !ok {
			return p.parseState.empty()
		}
		ok = p.beginVal(c)
	}
	return false
}

func (p *compactor) beginVal(c byte) bool {
	p.keep(c)

	switch c {
	case '{':
		p.parseState.push(parseObjKey)
		return p.beginStrOrEmpty()

	case '[':
		p.parseState.push(parseArrVal)
		return p.beginValOrEmpty()

	case '"':
		return p.beganStrFin()

	case '-':
		return p.finNeg()

	case '0':
		return p.fin0()

	case 't':
		return p.finTrue()

	case 'f':
		return p.finFalse()

	case 'n':
		return p.finNull()
	}
	if isNat(c) {
		return p.fin1()
	}
	return false
}

func (p *compactor) beginValOrEmpty() bool {
	c, ok := p.peekAfterSpace()
	if !ok {
		return false
	}
	if c == ']' {
		return p.endVal()
	}
	return true
}

func (p *compactor) beginStrOrEmpty() bool {
	c, ok := p.peekAfterSpace()
	if !ok {
		return false
	}
	if c == '}' {
		p.parseState.replace(parseObjVal)
		return p.endVal()
	}
	p.keepskip()
	return c == '"' && p.finStr() && p.endVal()
}

func (p *compactor) beganStrFin() bool {
	return p.finStr() && p.endVal()
}

func (p *compactor) finStr() bool {
start:
	c, ok := p.next()
	// we must wait to keep this; see blow
	if !ok {
		return false
	}
	if c == '"' {
		p.keep(c)
		return true
	}
	if c == '\\' {
		p.keep(c)
		if ok = p.finStrEsc(); !ok {
			return false
		}
		goto start
	}
	if c < 0x20 {
		return false
	}

	if c != 0xe2 || !p.jsonp {
		p.keep(c)
		goto start
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
	n1, n2, ok := p.peek2()
	if n1 != 0x80 || n2&^1 != 0xa8 {
		p.keep(c)
		goto start
	}

	srcPtr := uintptr(unsafe.Pointer(&p.in[p.at-1])) // sub one to get where we _were_
	dstPtr := uintptr(unsafe.Pointer(&p.dst[len(p.dst)-1]))

	// dstPtr, at most, can be just before srcPtr. We have not appended the
	// new character yet.
	//
	// We need to reallocate if our three new characters will put us on
	// or after srcPtr.
	if dstPtr < srcPtr && dstPtr+3 >= srcPtr {
		new := make([]byte, len(p.dst), cap(p.in))
		copy(new, p.dst)
		p.dst = new
	}

	p.dst = append(p.dst, `\u202`...)
	p.keep('8' | n2&1)
	p.at += 2

	goto start
}

func (p *compactor) finStrEsc() bool {
	c, ok := p.next()
	if !ok {
		return false
	}
	p.keep(c)
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		return true
	case 'u':
		for i := 0; i < 4; i++ {
			if c, ok = p.next(); !ok || !isHex(c) {
				return false
			}
			p.keep(c)
		}
		return true
	}
	return false
}

func (p *compactor) finNeg() bool {
	c, ok := p.next()
	if !ok {
		return false
	}
	p.keep(c)
	if c == '0' {
		return p.fin0()
	}
	if isNat(c) {
		return p.fin1()
	}
	return false
}

func (p *compactor) fin1() bool {
	for !p.done() {
		if !isNum(p.on()) {
			return p.fin0()
		}
		p.keepskip()
	}
	return p.endVal()
}

func (p *compactor) fin0() bool {
	c, ok := p.peek()
	if !ok {
		return p.endVal()
	}
	if c == '.' {
		p.keepskip()
		return p.finDot()
	}
	if isE(c) {
		p.keepskip()
		return p.finE()
	}
	return p.endVal()
}

func (p *compactor) finDot() bool {
	c, ok := p.next()
	if !ok || !isNum(c) { // first char after dot must be num
		return false
	}
	p.keep(c)

	for !p.done() && isNum(p.on()) { // consume all nums
		p.keepskip()
	}

	if !p.done() && isE(p.on()) {
		p.keepskip()
		return p.finE()
	}
	return p.endVal()
}

func (p *compactor) finE() bool {
	c, ok := p.next()
	if !ok {
		return false
	}
	if c == '+' || c == '-' {
		p.keep(c) // keep before overwriting c
		if c, ok = p.next(); !ok {
			return false
		}
	}
	if !isNum(c) { // first after e (and +/-) must be num
		return false
	}

	p.keep(c)

	for !p.done() && isNum(p.on()) { // consume all nums
		p.keepskip()
	}
	return p.endVal()
}

func (p *compactor) finTrue() bool {
	return p.trykeep3('r', 'u', 'e') && p.endVal()
}

func (p *compactor) finNull() bool {
	return p.trykeep3('u', 'l', 'l') && p.endVal()
}

func (p *compactor) finFalse() bool {
	return p.trykeep4('a', 'l', 's', 'e') && p.endVal()
}

func (p *compactor) endVal() bool {
start:
	if p.parseState.empty() {
		return p.remSpace()
	}

	c, ok := p.afterSpace()
	if !ok { // if parseState is not empty, we need another character
		return false
	}
	p.keep(c)

	switch p.parseState.pop() {
	case parseObjKey:
		if c == ':' {
			p.parseState.push(parseObjVal)
			return true
		}

	case parseObjVal:
		switch c {
		case ',':
			p.parseState.push(parseObjKey)
			if !p.startAndFinStr() {
				return false
			}
			goto start
		case '}':
			goto start
		}

	case parseArrVal:
		switch c {
		case ',':
			p.parseState.push(parseArrVal)
			return true
		case ']':
			goto start
		}
	}

	return false
}

func (p *compactor) startAndFinStr() bool {
	c, ok := p.afterSpace()
	p.keep(c)
	return ok && c == '"' && p.finStr()
}
