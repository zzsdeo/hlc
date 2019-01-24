package jstream

import (
	"io"
	"strconv"
	"unicode/utf16"
)

// MetaValue wraps a decoded interface value with the document
// position and depth at which the value was parsed
type MetaValue struct {
	Offset int
	Length int
	Depth  int
	Value  interface{}
}

// KV contains a key and value pair parsed from a decoded object
type KV struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Decoder wraps an io.Reader to provide incremental decoding of
// JSON values
type Decoder struct {
	*scanner
	emitDepth     int
	emitKV        bool
	emitRecursive bool

	depth   int
	scratch *scratch
	metaCh  chan *MetaValue
	err     error

	// follow line position to add context to errors
	lineNo    int
	lineStart int
}

// NewDecoder creates new Decoder to read JSON values at the provided
// emitDepth from the provider io.Reader.
// If emitDepth is < 0, values at every depth will be emitted.
func NewDecoder(r io.Reader, emitDepth int) *Decoder {
	d := &Decoder{
		scanner:   newScanner(r),
		emitDepth: emitDepth,
		scratch:   &scratch{data: make([]byte, 1024)},
		metaCh:    make(chan *MetaValue, 128),
	}
	if emitDepth < 0 {
		d.emitDepth = 0
		d.emitRecursive = true
	}
	return d
}

// EmitKV enables emitting a jstream.KV struct when the items(s) parsed
// at configured emit depth are within a JSON object. By default, only
// the object values are emitted.
func (d *Decoder) EmitKV() *Decoder {
	d.emitKV = true
	return d
}

// Recursive enables emitting all values at a depth higher than the
// configured emit depth; e.g. if an array is found at emit depth, all
// values within the array are emitted to the stream, then the array
// containing those values is emitted.
func (d *Decoder) Recursive() *Decoder {
	d.emitRecursive = true
	return d
}

// Stream begins decoding from the underlying reader and returns a
// streaming MetaValue channel for JSON values at the configured emitDepth.
func (d *Decoder) Stream() chan *MetaValue {
	go d.decode()
	return d.metaCh
}

// Pos returns the number of bytes consumed from the underlying reader
func (d *Decoder) Pos() int { return d.pos }

// Err returns the most recent decoder error if any, or nil
func (d *Decoder) Err() error { return d.err }

// Decode parses the JSON-encoded data and returns an interface value
func (d *Decoder) decode() {
	defer close(d.metaCh)
	d.skipSpaces()
	for d.pos < d.end {
		_, err := d.emitAny()
		if err != nil {
			d.err = err
			break
		}
		d.skipSpaces()
	}
}

func (d *Decoder) emitAny() (interface{}, error) {
	if d.pos >= d.end {
		return nil, d.mkError(ErrUnexpectedEOF)
	}
	offset := d.pos - 1
	i, err := d.any()
	if d.willEmit() {
		d.metaCh <- &MetaValue{
			Offset: offset,
			Length: d.pos - offset,
			Depth:  d.depth,
			Value:  i,
		}
	}
	return i, err
}

// return whether, at the current depth, the value being decoded will
// be emitted to stream
func (d *Decoder) willEmit() bool {
	if d.emitRecursive {
		return d.depth >= d.emitDepth
	}
	return d.depth == d.emitDepth
}

// any used to decode any valid JSON value, and returns an
// interface{} that holds the actual data
func (d *Decoder) any() (interface{}, error) {
	c := d.cur()

	switch c {
	case '"':
		return d.string()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.number()
	case '-':
		if c = d.next(); c < '0' && c > '9' {
			return nil, d.mkError(ErrSyntax, "in negative numeric literal")
		}
		n, err := d.number()
		if err != nil {
			return nil, err
		}
		return -n, nil
	case 'f':
		if d.remaining() < 4 {
			return nil, d.mkError(ErrUnexpectedEOF)
		}
		if d.next() == 'a' && d.next() == 'l' && d.next() == 's' && d.next() == 'e' {
			return false, nil
		}
		return nil, d.mkError(ErrSyntax, "in literal false")
	case 't':
		if d.remaining() < 3 {
			return nil, d.mkError(ErrUnexpectedEOF)
		}
		if d.next() == 'r' && d.next() == 'u' && d.next() == 'e' {
			return true, nil
		}
		return nil, d.mkError(ErrSyntax, "in literal true")
	case 'n':
		if d.remaining() < 3 {
			return nil, d.mkError(ErrUnexpectedEOF)
		}
		if d.next() == 'u' && d.next() == 'l' && d.next() == 'l' {
			return nil, nil
		}
		return nil, d.mkError(ErrSyntax, "in literal null")
	case '[':
		return d.array()
	case '{':
		return d.object()
	default:
		return nil, d.mkError(ErrSyntax, "looking for beginning of value")
	}
}

// string called by `any` or `object`(for map keys) after reading `"`
func (d *Decoder) string() (string, error) {
	d.scratch.reset()

	var (
		c = d.next()
	)

scan:
	for {
		switch {
		case c == '"':
			return string(d.scratch.bytes()), nil
		case c == '\\':
			c = d.next()
			goto scan_esc
		case c < 0x20:
			return "", d.mkError(ErrSyntax, "in string literal")
		// Coerce to well-formed UTF-8.
		default:
			d.scratch.add(c)
			if d.remaining() == 0 {
				return "", d.mkError(ErrSyntax, "in string literal")
			}
			c = d.next()
		}
	}

scan_esc:
	switch c {
	case '"', '\\', '/', '\'':
		d.scratch.add(c)
	case 'u':
		goto scan_u
	case 'b':
		d.scratch.add('\b')
	case 'f':
		d.scratch.add('\f')
	case 'n':
		d.scratch.add('\n')
	case 'r':
		d.scratch.add('\r')
	case 't':
		d.scratch.add('\t')
	default:
		return "", d.mkError(ErrSyntax, "in string escape code")
	}
	c = d.next()
	goto scan

scan_u:
	r := d.u4()
	if r < 0 {
		return "", d.mkError(ErrSyntax, "in unicode escape sequence")
	}

	// check for proceeding surrogate pair
	c = d.next()
	if !utf16.IsSurrogate(r) || c != '\\' {
		d.scratch.addRune(r)
		goto scan
	}
	if c = d.next(); c != 'u' {
		d.scratch.addRune(r)
		goto scan_esc
	}

	r2 := d.u4()
	if r2 < 0 {
		return "", d.mkError(ErrSyntax, "in unicode escape sequence")
	}

	// write surrogate pair
	d.scratch.addRune(utf16.DecodeRune(r, r2))
	c = d.next()
	goto scan
}

// u4 reads four bytes following a \u escape
func (d *Decoder) u4() rune {
	// logic taken from:
	// github.com/buger/jsonparser/blob/master/escape.go#L20
	var h [4]int
	for i := 0; i < 4; i++ {
		c := d.next()
		switch {
		case c >= '0' && c <= '9':
			h[i] = int(c - '0')
		case c >= 'A' && c <= 'F':
			h[i] = int(c - 'A' + 10)
		case c >= 'a' && c <= 'f':
			h[i] = int(c - 'a' + 10)
		default:
			return -1
		}
	}
	return rune(h[0]<<12 + h[1]<<8 + h[2]<<4 + h[3])
}

// number called by `any` after reading number between 0 to 9
func (d *Decoder) number() (float64, error) {
	d.scratch.reset()

	var (
		c       = d.cur()
		n       float64
		isFloat bool
	)

	// digits first
	switch {
	case c == '0':
		d.scratch.add(c)
		c = d.next()
	case '1' <= c && c <= '9':
		for ; c >= '0' && c <= '9'; c = d.next() {
			n = 10*n + float64(c-'0')
			d.scratch.add(c)
		}
	}

	// . followed by 1 or more digits
	if c == '.' {
		isFloat = true
		d.scratch.add(c)

		// first char following must be digit
		if c = d.next(); c < '0' && c > '9' {
			return 0, d.mkError(ErrSyntax, "after decimal point in numeric literal")
		}
		d.scratch.add(c)

		for {
			if d.remaining() == 0 {
				return 0, d.mkError(ErrUnexpectedEOF)
			}
			if c = d.next(); c < '0' || c > '9' {
				break
			}
			d.scratch.add(c)
		}
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if c == 'e' || c == 'E' {
		isFloat = true
		d.scratch.add(c)

		if c = d.next(); c == '+' || c == '-' {
			d.scratch.add(c)
			if c = d.next(); c < '0' || c > '9' {
				return 0, d.mkError(ErrSyntax, "in exponent of numeric literal")
			}
			d.scratch.add(c)
		}
		for ; c >= '0' && c <= '9'; c = d.next() {
			d.scratch.add(c)
		}
	}

	if isFloat {
		var (
			err error
			sn  string
		)
		sn = string(d.scratch.bytes())
		if n, err = strconv.ParseFloat(sn, 64); err != nil {
			return 0, err
		}
	}

	d.back()
	return n, nil
}

// array accept valid JSON array value
func (d *Decoder) array() ([]interface{}, error) {
	d.depth++

	var (
		c     byte
		v     interface{}
		err   error
		array = make([]interface{}, 0)
	)

	// look ahead for ] - if the array is empty.
	if c = d.skipSpaces(); c == ']' {
		goto out
	}

scan:
	if v, err = d.emitAny(); err != nil {
		goto out
	}

	if d.depth > d.emitDepth { // skip alloc for array if it won't be emitted
		array = append(array, v)
	}

	// next token must be ',' or ']'
	switch c = d.skipSpaces(); c {
	case ',':
		d.skipSpaces()
		goto scan
	case ']':
		goto out
	default:
		err = d.mkError(ErrSyntax, "after array element")
	}

out:
	d.depth--
	return array, err
}

// object accept valid JSON array value
func (d *Decoder) object() (map[string]interface{}, error) {
	d.depth++

	var (
		c   byte
		k   string
		v   interface{}
		err error
		obj map[string]interface{}
	)

	// skip allocating map if it will not be emitted
	if d.depth > d.emitDepth {
		obj = make(map[string]interface{})
	}

	// if the object has no keys
	if c = d.skipSpaces(); c == '}' {
		goto out
	}

scan:
	for {
		offset := d.pos - 1

		// read string key
		if c != '"' {
			err = d.mkError(ErrSyntax, "looking for beginning of object key string")
			break
		}
		if k, err = d.string(); err != nil {
			break
		}

		// read colon before value
		if c = d.skipSpaces(); c != ':' {
			err = d.mkError(ErrSyntax, "after object key")
			break
		}

		// read value
		d.skipSpaces()
		if d.emitKV {
			if v, err = d.any(); err != nil {
				break
			}
			if d.willEmit() {
				d.metaCh <- &MetaValue{
					Offset: offset,
					Length: d.pos - offset,
					Depth:  d.depth,
					Value:  KV{k, v},
				}
			}
		} else {
			if v, err = d.emitAny(); err != nil {
				break
			}
		}

		if obj != nil {
			obj[k] = v
		}

		// next token must be ',' or '}'
		switch c = d.skipSpaces(); c {
		case '}':
			goto out
		case ',':
			c = d.skipSpaces()
			goto scan
		default:
			err = d.mkError(ErrSyntax, "after object key:value pair")
		}
	}

out:
	d.depth--
	return obj, err
}

// returns the next char after white spaces
func (d *Decoder) skipSpaces() byte {
	for d.pos < d.end {
		switch c := d.next(); c {
		case '\n':
			d.lineStart = d.pos
			d.lineNo++
			continue
		case ' ', '\t', '\r':
			continue
		default:
			return c
		}
	}
	return 0
}

// create syntax errors at current position, with optional context
func (d *Decoder) mkError(err SyntaxError, context ...string) error {
	if len(context) > 0 {
		err.context = context[0]
	}
	err.atChar = d.cur()
	err.pos[0] = d.lineNo + 1
	err.pos[1] = d.pos - d.lineStart
	return err
}
