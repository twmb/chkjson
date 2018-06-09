// Package chkjson checks whether data is valid JSON and compacts it
// (optionally in-place) without allocating.
//
// The standard library allocates when validating JSON and allocates more if
// the JSON is very nested. This library only allocates if the JSON is more
// than 32 levels deep (objects and arrays). Other minor to major differences
// allow this library's Valid to over 2x faster as encoding/json's.
//
// The standard library ensures that compact JSON is JavaScript safe. This is
// necessary if the JSON will ever end up in JSONP, but is not always
// necessary. This library provides a faster AppendCompactJSONP function to
// imitate the stdlib's Compact, and further provides AppendCompact for the
// same speed with a no-allocation (no JSONP escaping) guarantee. This
// library's AppendCompact is usually around 4x faster than encoding/json's
// Compact.
//
// In essence, this library aims to provide faster and allocation free
// alternatives to encoding/json for a few specific use cases.
package chkjson

type parseState byte

const (
	parseObjKey parseState = iota
	parseObjVal
	parseArrVal
)

// parseStateStack contains the knowledge of how nested we are in { or [.
//
// We avoid allocations by providing a 32 byte default stack; most JSON (that I
// at least deal with) is not so deep, which allows us to never allocate an
// extension slice.
type parseStateStack struct {
	base [32]parseState
	end  uint8
	ext  []parseState
}

func (p *parseStateStack) pop() parseState { // faster to not inline in endVal
	l := len(p.ext)
	if l == 0 {
		r := p.base[p.end-1]
		p.end--
		return r
	}

	r := p.ext[l-1]
	p.ext = p.ext[:l-1]
	return r
}

func (p *parseStateStack) empty() bool {
	return p.end == 0
}

type parser struct {
	in []byte
	at int

	stack parseStateStack
}

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
// Our scanner instead bottoms out evaluating an entire value at a time. This
// avoids recursion because every value has a defined limited bottom, but also
// allows us to avoid setting the next function to use for evaluating. Every
// "next function" is the beginning of a value. Whenever we begin a value, we
// know what the rest of it must parse as.
//
// We use a giant state-machine function and inline many functions for
// significant performance gain (on the order of 50% faster).

// Valid returns whether b is valid JSON.
func Valid(b []byte) bool {
	p := parser{in: b}
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
		goto endVal
	}
	if c == '\\' { // finStrEsc
		if p.at >= len(p.in) {
			return false
		}
		c = p.in[p.at]
		p.at++
		switch c {
		case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
			goto finStr
		case 'u':
			if p.at+3 < len(p.in) &&
				isHex(p.in[p.at]) &&
				isHex(p.in[p.at+1]) &&
				isHex(p.in[p.at+2]) &&
				isHex(p.in[p.at+3]) {
				p.at += 4
				goto finStr
			}
		}
		return false
	}
	if c < 0x20 {
		return false
	}
	goto finStr

finNeg:
	if p.at >= len(p.in) {
		return false
	}
	c = p.in[p.at]
	p.at++
	if c == '0' {
		goto fin0
	}
	if isNat(c) {
		goto fin1
	}
	return false

fin1:
	for p.at < len(p.in) {
		if !isNum(p.in[p.at]) {
			goto fin0
		}
		p.at++
	}
	goto endVal

fin0:
	if p.at >= len(p.in) {
		goto endVal
	}
	c = p.in[p.at]
	if c == '.' {
		p.at++
		goto finDot
	}
	if isE(c) {
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

	for p.at < len(p.in) && isNum(p.in[p.at]) { // consume all nums
		p.at++
	}

	if p.at < len(p.in) && isE(p.in[p.at]) {
		p.at++
		goto finE
	}
	goto endVal

finE:
	if p.at >= len(p.in) {
		return false
	}
	c = p.in[p.at]
	p.at++
	if c == '+' || c == '-' {
		if p.at >= len(p.in) {
			return false
		}
		c = p.in[p.at]
		p.at++
	}
	if !isNum(c) { // first after e (and +/-) must be num
		return false
	}
	for p.at < len(p.in) && isNum(p.in[p.at]) { // consume all nums
		p.at++
	}
	goto endVal

finTrue:
	if p.at+2 < len(p.in) &&
		p.in[p.at] == 'r' &&
		p.in[p.at+1] == 'u' &&
		p.in[p.at+2] == 'e' {

		p.at += 3
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
		goto endVal
	}
	return false

finNull:
	if p.at+2 < len(p.in) &&
		p.in[p.at] == 'u' &&
		p.in[p.at+1] == 'l' &&
		p.in[p.at+2] == 'l' {

		p.at += 3
		goto endVal
	}
	return false

endVal: // most things parsed fall into endVal
	if p.stack.empty() {
		for p.at < len(p.in) { // afterSpace
			c = p.in[p.at]
			p.at++
			if !isSpace(c) {
				return false
			}
		}
		return true // the rest are space if we have nothing after spaces
	}

	for p.at < len(p.in) { // if parseState is not empty, we need another character
		c = p.in[p.at]
		p.at++
		if !isSpace(c) {
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

func isSpace(c byte) bool {
	return c <= ' ' && (c == ' ' || c == '\t' || c == '\r' || c == '\n')
}

func isHex(c byte) bool {
	return isNum(c) || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

func isNum(c byte) bool {
	return '0' <= c && c <= '9'
}

func isNat(c byte) bool {
	return '1' <= c && c <= '9'
}

func isE(c byte) bool {
	return c == 'e' || c == 'E'
}
