// Package chkjson checks whether data is valid JSON and compacts it
// (optionally in-place) without allocating.
//
// The standard library allocates when validating JSON and allocates more if
// the JSON is very nested. This library only allocates if the JSON is more
// than 32 levels deep (objects and arrays). Other minor differences allow this
// library's Valid to be anywhere from 5% to 25% faster than encoding/json.
//
// The standard library ensures that compact JSON is JavaScript safe. This is
// necessary if the JSON will ever end up in JSONP, but is not always
// necessary. This library provides a faster AppendCompactJSONP function to
// imitate the stdlib's Compact, and further provides AppendCompact for the
// same speed with a no-allocation guarantee. This library's AppendCompact is
// around 50% faster than encoding/json's Compact.
//
// In essence, this library aims to provide slightly faster and allocation free
// alternatives to encoding/json for a few specific use cases.
package chkjson

// Valid returns whether b is valid JSON.
func Valid(b []byte) bool {
	p := parser{in: b}
	_, ok := p.peekAfterSpace()
	if !ok {
		return false
	}
	return p.validate()
}

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
	stackBase [32]parseState
	baseEnd   uint8
	stackExt  []parseState
}

func (p *parseStateStack) push(s parseState) {
	if p.baseEnd < 32 {
		p.stackBase[p.baseEnd] = s
		p.baseEnd++
		return
	}
	p.stackExt = append(p.stackExt, s)
}

func (p *parseStateStack) replace(s parseState) {
	l := len(p.stackExt)
	if l == 0 {
		p.stackBase[p.baseEnd-1] = s
		return
	}

	p.stackExt[l-1] = s
}

func (p *parseStateStack) pop() parseState {
	l := len(p.stackExt)
	if l == 0 {
		r := p.stackBase[p.baseEnd-1]
		p.baseEnd--
		return r
	}

	r := p.stackExt[l-1]
	p.stackExt = p.stackExt[:l-1]
	return r
}

func (p *parseStateStack) empty() bool {
	return p.baseEnd == 0
}

type parser struct {
	in []byte
	at int // remove in go1.11: https://go-review.googlesource.com/c/go/+/105258

	parseState parseStateStack
}

func (p *parser) done() bool { return p.at == len(p.in) }
func (p *parser) on() byte   { return p.in[p.at] }
func (p *parser) skip()      { p.at++ }
func (p *parser) c() byte    { c := p.on(); p.skip(); return c }

func (p *parser) peek() (byte, bool) {
	if p.done() {
		return 0, false
	}
	return p.on(), true
}

func (p *parser) peek2() (byte, byte, bool) {
	if p.at+2 <= len(p.in) {
		return p.in[p.at], p.in[p.at+1], true
	}
	return 0, 0, false
}

func (p *parser) next() (byte, bool) {
	if p.done() {
		return 0, false
	}
	return p.c(), true
}

func (p *parser) try3(c1, c2, c3 byte) bool {
	return p.at+3 <= len(p.in) &&
		p.c() == c1 &&
		p.c() == c2 &&
		p.c() == c3
}

func (p *parser) try4(c1, c2, c3, c4 byte) bool {
	return p.at+4 <= len(p.in) &&
		p.c() == c1 &&
		p.c() == c2 &&
		p.c() == c3 &&
		p.c() == c4
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func (p *parser) afterSpace() (byte, bool) {
	for {
		c, ok := p.next()
		if !ok {
			return 0, false
		}
		if !isSpace(c) {
			return c, true
		}
	}
}

func (p *parser) peekAfterSpace() (byte, bool) {
	for {
		c, ok := p.peek()
		if !ok {
			return 0, false
		}
		if !isSpace(c) {
			return c, true
		}
		p.skip()
	}
}

// validate validates JSON starting from the knowledge that some non-space
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
func (p *parser) validate() bool {
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

func (p *parser) beginVal(c byte) bool {
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

func (p *parser) beginValOrEmpty() bool {
	c, ok := p.peekAfterSpace()
	if !ok {
		return false
	}
	if c == ']' {
		return p.endVal()
	}
	return true
}

func (p *parser) beginStrOrEmpty() bool {
	c, ok := p.peekAfterSpace()
	if !ok {
		return false
	}
	if c == '}' {
		p.parseState.replace(parseObjVal)
		return p.endVal()
	}
	p.skip()
	return c == '"' && p.finStr() && p.endVal()
}

// beganStrFin finishes a started string and falls into endVal.
func (p *parser) beganStrFin() bool {
	return p.finStr() && p.endVal()
}

func (p *parser) finStr() bool {
start:
	c, ok := p.next()
	if !ok {
		return false
	}
	if c == '"' {
		return true
	}
	if c == '\\' {
		if ok = p.finStrEsc(); !ok {
			return false
		}
		goto start
	}
	if c < 0x20 {
		return false
	}
	goto start
}

func (p *parser) finStrEsc() bool {
	c, ok := p.next()
	if !ok {
		return false
	}
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		return true
	case 'u':
		for i := 0; i < 4; i++ {
			if c, ok = p.next(); !ok || !isHex(c) {
				return false
			}
		}
		return true
	}
	return false
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

func (p *parser) finNeg() bool {
	c, ok := p.next()
	if !ok {
		return false
	}
	if c == '0' {
		return p.fin0()
	}
	if isNat(c) {
		return p.fin1()
	}
	return false
}

func (p *parser) fin1() bool {
	for !p.done() {
		if !isNum(p.on()) {
			return p.fin0()
		}
		p.skip()
	}
	return p.endVal()
}

func (p *parser) fin0() bool {
	c, ok := p.peek()
	if !ok {
		return p.endVal()
	}
	if c == '.' {
		p.skip()
		return p.finDot()
	}
	if isE(c) {
		p.skip()
		return p.finE()
	}
	return p.endVal()
}

func (p *parser) finDot() bool {
	c, ok := p.next()
	if !ok || !isNum(c) { // first char after dot must be num
		return false
	}

	for !p.done() && isNum(p.on()) { // consume all nums
		p.skip()
	}

	if !p.done() && isE(p.on()) {
		p.skip()
		return p.finE()
	}
	return p.endVal()
}

func (p *parser) finE() bool {
	c, ok := p.next()
	if !ok {
		return false
	}
	if c == '+' || c == '-' {
		if c, ok = p.next(); !ok {
			return false
		}
	}
	if !isNum(c) { // first after e (and +/-) must be num
		return false
	}
	for !p.done() && isNum(p.on()) { // consume all nums
		p.skip()
	}
	return p.endVal()
}

func (p *parser) finTrue() bool {
	return p.try3('r', 'u', 'e') && p.endVal()
}

func (p *parser) finNull() bool {
	return p.try3('u', 'l', 'l') && p.endVal()
}

func (p *parser) finFalse() bool {
	return p.try4('a', 'l', 's', 'e') && p.endVal()
}

// endVal is the bottom of every value. All functions called from endVal cannot
// fall back into endVal.
func (p *parser) endVal() bool {
start:
	if p.parseState.empty() {
		return p.remSpace()
	}

	c, ok := p.afterSpace()
	if !ok { // if parseState is not empty, we need another character
		return false
	}

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

func (p *parser) remSpace() bool {
	_, ok := p.afterSpace()
	return !ok // the rest are space if we have nothing after spaces
}

// startAndFinStr begins and finishes a string; this does not fall into endVal,
// thus it can be called from endVal.
func (p *parser) startAndFinStr() bool {
	c, ok := p.afterSpace()
	return ok && c == '"' && p.finStr()
}
